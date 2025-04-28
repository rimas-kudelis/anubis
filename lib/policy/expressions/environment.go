package expressions

import (
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
)

// NewEnvironment creates a new CEL environment, this is the set of
// variables and functions that are passed into the CEL scope so that
// Anubis can fail loudly and early when something is invalid instead
// of blowing up at runtime.
func NewEnvironment() (*cel.Env, error) {
	return cel.NewEnv(
		ext.Strings(
			ext.StringsLocale("en_US"),
			ext.StringsValidateFormatCalls(true),
		),

		// default all timestamps to UTC
		cel.DefaultUTCTimeZone(true),

		// Variables exposed to CEL programs:

		// Request metadata
		cel.Variable("remoteAddress", cel.StringType),
		cel.Variable("host", cel.StringType),
		cel.Variable("method", cel.StringType),
		cel.Variable("userAgent", cel.StringType),
		cel.Variable("path", cel.StringType),
		cel.Variable("query", cel.MapType(cel.StringType, cel.StringType)),
		cel.Variable("headers", cel.MapType(cel.StringType, cel.StringType)),

		// System load metadata
		cel.Variable("load_1m", cel.DoubleType),
		cel.Variable("load_5m", cel.DoubleType),
		cel.Variable("load_15m", cel.DoubleType),

		// Functions exposed to CEL programs:

		// userAgent.isBrowserLike() method, used to detect if a user agent is likely a browser
		// based on shibboleth words in the User-Agent string.
		cel.Function("isBrowserLike",
			cel.MemberOverload("userAgent_isBrowserLike_string",
				[]*cel.Type{cel.StringType},
				cel.BoolType,
				cel.UnaryBinding(func(userAgentVal ref.Val) ref.Val {
					var userAgent string
					switch v := userAgentVal.Value().(type) {
					case string:
						userAgent = v
					default:
						return types.NewErr("invalid type %T", userAgentVal)
					}

					switch {
					case strings.Contains(userAgent, "Mozilla"), strings.Contains(userAgent, "Opera"), strings.Contains(userAgent, "Gecko"), strings.Contains(userAgent, "WebKit"), strings.Contains(userAgent, "Apple"), strings.Contains(userAgent, "Chrome"), strings.Contains(userAgent, "Windows"), strings.Contains(userAgent, "Linux"):
						return types.Bool(true)
					default:
						return types.Bool(false)
					}
				}),
			),
		),
	)
}

// Compile takes CEL environment and syntax tree then emits an optimized
// Program for execution.
func Compile(env *cel.Env, ast *cel.Ast) (cel.Program, error) {
	return env.Program(
		ast,
		cel.EvalOptions(
			// optimize regular expressions right now instead of on the fly
			cel.OptOptimize,
		),
	)
}
