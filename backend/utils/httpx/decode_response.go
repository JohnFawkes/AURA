package httpx

import (
	"aura/logging"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

func DecodeResponseToJSON(ctx context.Context, body []byte, out any, structName string) logging.LogErrorInfo {
	decoder := json.NewDecoder(bytes.NewReader(body))
	err := decoder.Decode(out)
	if err != nil {
		_, logAction := logging.AddSubActionToContext(ctx, fmt.Sprintf("Decoding response body into `%s` struct", structName), logging.LevelTrace)
		defer logAction.Complete()
		errorDetails := map[string]any{
			"response_body": string(body),
			"error":         err.Error(),
			"error_type":    fmt.Sprintf("%T", err),
		}

		// Enhanced error reporting for JSON errors
		switch e := err.(type) {
		case *json.UnmarshalTypeError:
			errorDetails["field"] = e.Field
			errorDetails["value"] = e.Value
			errorDetails["type"] = e.Type.String()
			// logAction.SetError(
			// 	fmt.Sprintf("Type error for field '%s' in struct '%s'", e.Field, structName),
			// 	fmt.Sprintf("Check the type of field '%s' (expected %s, got %s)", e.Field, e.Type.String(), e.Value),
			// 	errorDetails,
			// )
			logging.LOGGER.Error().Timestamp().Err(err).Msgf("Type error for field `%s` in struct `%s`: expected %s, got %s", e.Field, structName, e.Type.String(), e.Value)
		case *json.SyntaxError:
			errorDetails["offset"] = e.Offset
			// logAction.SetError(
			// 	"JSON syntax error",
			// 	fmt.Sprintf("Syntax error at offset %d in struct '%s'", e.Offset, structName),
			// 	errorDetails,
			// )
			logging.LOGGER.Error().Timestamp().Err(err).Msgf("JSON syntax error at offset %d in struct `%s`", e.Offset, structName)
		case *json.InvalidUnmarshalError:
			errorDetails["type"] = e.Type.String()
			// logAction.SetError(
			// 	"Invalid unmarshal error",
			// 	fmt.Sprintf("Invalid unmarshal to type %s in struct '%s'", e.Type.String(), structName),
			// 	errorDetails,
			// )
			logging.LOGGER.Error().Timestamp().Err(err).Msgf("Invalid unmarshal to type %s in struct `%s`", e.Type.String(), structName)
		case *json.UnsupportedTypeError:
			errorDetails["type"] = e.Type.String()
			// logAction.SetError(
			// 	"Unsupported type error",
			// 	fmt.Sprintf("Unsupported type %s in struct '%s'", e.Type.String(), structName),
			// 	errorDetails,
			// )
			logging.LOGGER.Error().Timestamp().Err(err).Msgf("Unsupported type %s in struct `%s`", e.Type.String(), structName)
		default:
			// logAction.SetError(
			// 	"Failed to decode the JSON response body",
			// 	fmt.Sprintf("Ensure that the JSON is correct for %s", structName),
			// 	errorDetails,
			// )
			logging.LOGGER.Error().Timestamp().Err(err).Msgf("Failed to decode JSON response into struct `%s`", structName)
		}
		return *logAction.Error
	}
	return logging.LogErrorInfo{}
}
