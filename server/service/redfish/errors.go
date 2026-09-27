package redfish

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// registry is the Base message registry the error messages come from.
const registry = "Base.1.16"

type message struct {
	text       string
	severity   string
	resolution string
}

// messages holds the Base registry entries the service answers with. %1, %2
// and %3 are replaced by the message arguments.
var messages = map[string]message{
	"GeneralError": {
		"A general error has occurred.  See Resolution for information on how to resolve the error.",
		"Critical", "None.",
	},
	"MalformedJSON": {
		"The request body submitted was malformed JSON and could not be parsed by the receiving service.",
		"Critical", "Ensure that the request body is valid JSON and resubmit the request.",
	},
	"ResourceMissingAtURI": {
		"The resource at the URI '%1' was not found.",
		"Critical", "Place a valid resource at the URI or correct the URI and resubmit the request.",
	},
	"ResourceNotFound": {
		"The requested resource of type '%1' named '%2' was not found.",
		"Critical", "Provide a valid resource identifier and resubmit the request.",
	},
	"ResourceInUse": {
		"The change to the requested resource failed because the resource is in use or in transition.",
		"Warning", "Remove the condition and resubmit the request if the operation failed.",
	},
	"NoValidSession": {
		"There is no valid session established with the implementation.",
		"Critical", "Establish a session before attempting any operations.",
	},
	"InsufficientPrivilege": {
		"There are insufficient privileges for the account or credentials associated with the current session to perform the requested operation.",
		"Critical", "Either abandon the operation or change the associated access rights and resubmit the request if the operation failed.",
	},
	"PropertyMissing": {
		"The property '%1' is a required property and must be included in the request.",
		"Warning", "Ensure that the property is in the request body and has a valid value and resubmit the request if the operation failed.",
	},
	"ActionNotSupported": {
		"The action '%1' is not supported by the resource.",
		"Critical", "The action supplied cannot be resubmitted to the implementation.  Perhaps the action was invalid, the wrong resource was the target or the implementation documentation may be of assistance.",
	},
	"ActionParameterMissing": {
		"The action '%1' requires the parameter '%2' to be present in the request body.",
		"Critical", "Supply the action with the required parameter in the request body when the request is resubmitted.",
	},
	"ActionParameterNotSupported": {
		"The parameter '%1' for the action '%2' is not supported on the target resource.",
		"Warning", "Remove the parameter supplied and resubmit the request if the operation failed.",
	},
	"ActionParameterValueNotInList": {
		"The value '%1' for the parameter '%2' in the action '%3' is not in the list of acceptable values.",
		"Warning", "Choose a value from the enumeration list that the implementation can support and resubmit the request if the operation failed.",
	},
	"ActionParameterValueTypeError": {
		"The value '%1' for the parameter '%2' in the action '%3' is of a different type than the parameter can accept.",
		"Warning", "Correct the value for the parameter in the request body and resubmit the request if the operation failed.",
	},
}

// writeError answers with a Redfish error object and stops the handler chain.
// detail, when set, is the top-level message; otherwise the registry text is.
func writeError(c *gin.Context, status int, id string, detail string, args ...string) {
	entry, ok := messages[id]
	if !ok {
		entry = messages["GeneralError"]
		id = "GeneralError"
	}

	text := entry.text
	for i, arg := range args {
		text = strings.ReplaceAll(text, "%"+string(rune('1'+i)), arg)
	}
	if detail == "" {
		detail = text
	}
	if args == nil {
		args = []string{}
	}

	writeJSON(c, status, object{
		"error": object{
			"code":    registry + "." + id,
			"message": detail,
			"@Message.ExtendedInfo": []object{{
				"@odata.type":     "#Message.v1_1_1.Message",
				"MessageId":       registry + "." + id,
				"Message":         text,
				"MessageArgs":     args,
				"MessageSeverity": entry.severity,
				"Resolution":      entry.resolution,
			}},
		},
	})
	c.Abort()
}

// maxBody bounds a request body. Every body the service takes is a handful
// of short properties.
const maxBody = 64 << 10

// decodeParams reads a JSON object body into its top-level properties. An
// empty body is an empty object, so a bare POST to EjectMedia works. On
// failure it has answered already and returns false.
func decodeParams(c *gin.Context) (map[string]json.RawMessage, bool) {
	data, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBody+1))
	if err != nil || len(data) > maxBody {
		writeError(c, http.StatusBadRequest, "GeneralError", "the request body is too large or could not be read")
		return nil, false
	}

	params := map[string]json.RawMessage{}
	if len(strings.TrimSpace(string(data))) == 0 {
		return params, true
	}

	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(c, http.StatusUnsupportedMediaType, "GeneralError", "the request body must be application/json")
		return nil, false
	}

	if err := json.Unmarshal(data, &params); err != nil || params == nil {
		writeError(c, http.StatusBadRequest, "MalformedJSON", "")
		return nil, false
	}

	return params, true
}

// present reports whether a property was sent with a value other than null.
func present(params map[string]json.RawMessage, name string) bool {
	raw, ok := params[name]
	return ok && strings.TrimSpace(string(raw)) != "null"
}
