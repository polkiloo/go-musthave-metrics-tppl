package commoncfg

import (
	"errors"
	"flag"
	"fmt"
	"testing"
)

type agg struct {
	A     string
	B     int
	C     bool
	count int
}

type aVal struct{ S string }
type bVal struct{ N int }
type cVal struct{ On bool }
type unknownVal struct{}

func parseA(val string, present bool) (*aVal, error) {
	if !present {
		return nil, nil
	}
	return &aVal{S: val}, nil
}
func parseB(val string, present bool) (*bVal, error) {
	if !present {
		return nil, nil
	}
	var n int
	if _, err := fmt.Sscanf(val, "%d", &n); err != nil {
		return nil, err
	}
	return &bVal{N: n}, nil
}
func parseC(val string, present bool) (*cVal, error) {
	if !present {
		return nil, nil
	}
	return &cVal{On: val == "true"}, nil
}

func applyToAgg(dst *agg, v FlagValue) error {
	switch t := v.(type) {
	case nil:
		return nil
	case *aVal:
		if t != nil {
			dst.A = t.S
			dst.count++
		}
	case *bVal:
		if t != nil {
			dst.B = t.N
			dst.count++
		}
	case *cVal:
		if t != nil {
			dst.C = t.On
			dst.count++
		}
	case *unknownVal:
		return errors.New("apply fail")
	default:
		return fmt.Errorf("unsupported type %T", t)
	}
	return nil
}

func TestDispatcher_MergesHandlers_AllOK(t *testing.T) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.String("a", "def", "string flag")
	fs.Int("b", 0, "int flag")
	fs.Bool("c", false, "bool flag")

	out, err := NewDispatcher[agg](fs, applyToAgg).
		Handle("a", Lift(parseA)).
		Handle("b", Lift(parseB)).
		Handle("c", Lift(parseC)).
		Parse([]string{"-a", "hello", "-b=42", "-c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.A != "hello" || out.B != 42 || out.C != true {
		t.Fatalf("merged mismatch: %+v", out)
	}
	if out.count != 3 {
		t.Fatalf("want 3 applications, got %d", out.count)
	}
}

func TestDispatcher_PresentFalse_NotApplied(t *testing.T) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.String("a", "def", "string flag")
	fs.Int("b", 0, "int flag")
	fs.Bool("c", false, "bool flag")

	out, err := NewDispatcher[agg](fs, applyToAgg).
		Handle("a", Lift(parseA)).
		Handle("b", Lift(parseB)).
		Handle("c", Lift(parseC)).
		Parse([]string{"-b", "7"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.A != "" || out.B != 7 || out.C != false {
		t.Fatalf("unexpected result: %+v", out)
	}
	if out.count != 1 {
		t.Fatalf("want 1 application, got %d", out.count)
	}
}

func TestDispatcher_PositionalArgs_Error(t *testing.T) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.String("a", "def", "string flag")

	_, err := NewDispatcher[agg](fs, applyToAgg).
		Handle("a", Lift(parseA)).
		Parse([]string{"-a", "x", "positional"})
	if !errors.Is(err, ErrUnknownArgs) {
		t.Fatalf("want ErrUnknownArgs, got %v", err)
	}
}

func TestDispatcher_UnknownFlag_ParseError(t *testing.T) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.String("a", "def", "string flag")

	_, err := NewDispatcher[agg](fs, applyToAgg).
		Handle("a", Lift(parseA)).
		Parse([]string{"-x"})
	if err == nil {
		t.Fatalf("expected parse error for unknown flag")
	}
}

func TestDispatcher_HandlerError_Propagates(t *testing.T) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.String("a", "def", "string flag")

	myErr := errors.New("boom")
	bad := func(_ string, present bool) (*aVal, error) {
		if present {
			return nil, myErr
		}
		return nil, nil
	}

	_, err := NewDispatcher[agg](fs, applyToAgg).
		Handle("a", Lift(bad)).
		Parse([]string{"-a", "v"})
	if !errors.Is(err, myErr) {
		t.Fatalf("want handler error %v, got %v", myErr, err)
	}
}

func TestDispatcher_ApplyError_Propagates(t *testing.T) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.String("u", "", "unknown value flag")

	parseUnknown := func(_ string, present bool) (*unknownVal, error) {
		if !present {
			return nil, nil
		}
		return &unknownVal{}, nil
	}

	_, err := NewDispatcher[agg](fs, applyToAgg).
		Handle("u", Lift(parseUnknown)).
		Parse([]string{"-u", "x"})
	if err == nil {
		t.Fatalf("expected apply error, got nil")
	}
}

func TestDispatcher_UnregisteredHandler_Ignored(t *testing.T) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.String("a", "def", "string flag")

	out, err := NewDispatcher[agg](fs, applyToAgg).
		Handle("a", Lift(parseA)).
		Handle("z", Lift(parseB)).
		Parse([]string{"-a", "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.A != "hi" || out.B != 0 || out.count != 1 {
		t.Fatalf("unexpected result with unregistered handler: %+v", out)
	}
}

func TestDispatcher_InlineForms_OK(t *testing.T) {
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	fs.String("a", "def", "string flag")
	fs.Int("b", 0, "int flag")
	fs.Bool("c", false, "bool flag")

	out, err := NewDispatcher[agg](fs, applyToAgg).
		Handle("a", Lift(parseA)).
		Handle("b", Lift(parseB)).
		Handle("c", Lift(parseC)).
		Parse([]string{"-a=hello", "-b=5", "-c=true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.A != "hello" || out.B != 5 || out.C != true {
		t.Fatalf("inline mismatch: %+v", out)
	}
	if out.count != 3 {
		t.Fatalf("want 3 applications, got %d", out.count)
	}
}
