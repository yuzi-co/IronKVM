package hostdevices

import (
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"

	"NanoKVM-Server/service/hostpower"
)

// The tool descriptions and input schemas. The callers name their tools
// differently, so a description that points to another tool takes its name.

// PowerDescription describes the press tool. statusTool is the caller's power
// status tool.
func PowerDescription(statusTool string) string {
	return "Press the ATX power or reset button of the remote host. " +
		"A short power press asks the operating system to shut down, or switches on a host that is off; " +
		"a long power press holds the button for 5 seconds and forces the host off at once, losing unsaved work. " +
		"A short reset press restarts the host immediately, also losing unsaved work. " +
		"Check " + statusTool + " first: the same short press switches a host on or off depending on its state. " +
		"Only one press runs at a time; a call made while another press is held is refused."
}

// PowerStatusDescription describes the power status tool.
const PowerStatusDescription = "Report whether the remote host is powered on, by its power LED: " +
	"state is on, off or unknown. Unknown means the LED header is not wired to the NanoKVM or could not be read, " +
	"so the power state cannot be told; use a screenshot instead and do not assume the host is off."

// MediaListDescription describes the media listing tool.
const MediaListDescription = "List the virtual drives the remote host sees (cdrom and disk), " +
	"the image each one holds (empty when none) and whether it is read-only, " +
	"and the image files on the NanoKVM that can be inserted."

// MediaDescription describes the insert and eject tool. listTool is the
// caller's media listing tool.
func MediaDescription(listTool string) string {
	return "Insert an image into a virtual drive of the remote host, or eject it. " +
		"Use a file path exactly as " + listTool + " returns it. Inserting replaces any image already in the drive. " +
		"The cdrom drive is always read-only. The disk is read-only unless readOnly is false; " +
		"a writable disk lets the host change or format the image, so only ask for it when the user wants that. " +
		"Ejecting an image the host is using can lose its writes. " +
		"On a board that supports it, insert can take a url instead of a file: the image is read from that " +
		"http or https server on demand, read-only, without being stored; the server must support byte ranges."
}

// PowerSchema is the press tool's input. withHold offers holdMs, an exact
// hold time, besides short and long.
func PowerSchema(withHold bool) *jsonschema.Schema {
	properties := map[string]*jsonschema.Schema{
		"button": {Type: "string", Enum: []any{hostpower.ButtonPower, hostpower.ButtonReset}, Description: "The front-panel button to press"},
		"press": {Type: "string", Enum: []any{PressShort, PressLong}, Description: fmt.Sprintf(
			"short holds the button %s (the default); long holds it %s, which on the power button forces the host off",
			hostpower.ShortPress, hostpower.LongPress)},
	}
	if withHold {
		properties["holdMs"] = &jsonschema.Schema{
			Type:        "integer",
			Description: "Exact hold time in milliseconds, instead of press",
			Minimum:     jsonschema.Ptr(float64(MinHoldMS)),
			Maximum:     jsonschema.Ptr(float64(MaxHoldMS)),
		}
	}
	return ObjectSchema(properties, "button")
}

// MediaSchema is the insert and eject tool's input. listTool is the caller's
// media listing tool.
func MediaSchema(listTool string) *jsonschema.Schema {
	return ObjectSchema(map[string]*jsonschema.Schema{
		"action":   {Type: "string", Enum: []any{"insert", "eject"}},
		"drive":    {Type: "string", Enum: []any{"cdrom", "disk"}, Description: "The drive to insert into or eject"},
		"file":     {Type: "string", Description: "insert: an image path exactly as " + listTool + " returns it", MaxLength: jsonschema.Ptr(512)},
		"readOnly": {Type: "boolean", Description: "insert into disk: false lets the host write to and format the image; defaults to true. The cdrom drive is always read-only."},
		"url":      {Type: "string", Description: "insert: an http or https URL to serve the image from on demand, read-only, instead of a file", MaxLength: jsonschema.Ptr(2048)},
	}, "action", "drive")
}

// ObjectSchema is an object with the properties and no others.
func ObjectSchema(properties map[string]*jsonschema.Schema, required ...string) *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:                 "object",
		Properties:           properties,
		Required:             required,
		AdditionalProperties: &jsonschema.Schema{Not: &jsonschema.Schema{}},
	}
}
