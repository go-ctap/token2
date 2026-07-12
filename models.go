package token2

import (
	"slices"
	"strconv"
	"strings"
)

// Model describes a Token2 hardware model identified by either a serial-number
// prefix or an inclusive serial-number range.
type Model struct {
	Revision   string
	FormFactor string
	Branding   string
	Prefix     string

	SerialFrom uint64
	SerialTo   uint64
}

// DisplayName returns the canonical human-readable name of the model.
func (m Model) DisplayName() string {
	parts := make([]string, 0, 3)
	for _, part := range []string{m.Branding, m.FormFactor, m.Revision} {
		if part = strings.TrimSpace(part); part != "" {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, " ")
}

var models = []Model{
	{Revision: "R1", FormFactor: "USB-A NFC", Branding: "Token2", Prefix: "86105"},
	{Revision: "R1", FormFactor: "USB-C NFC", Branding: "Token2", Prefix: "86104"},
	{Revision: "R1", FormFactor: "Dual NFC", Branding: "Token2", Prefix: "86103"},
	{Revision: "R1", FormFactor: "FIDO Card", Branding: "Token2", Prefix: "86202"},

	{Revision: "R2", FormFactor: "USB-A PIN+ NFC", Branding: "Token2", Prefix: "96105"},
	{Revision: "R2", FormFactor: "USB-C PIN+ NFC", Branding: "Token2", Prefix: "96104"},
	{Revision: "R2", FormFactor: "Dual PIN+ NFC", Branding: "Token2", Prefix: "96103"},
	{Revision: "R2", FormFactor: "Dual PIN+ NFC", Branding: "Unbranded", Prefix: "23103"},

	{Revision: "R3", FormFactor: "Dual PIN+ NFC", Branding: "Token2", Prefix: "76103"},
	{Revision: "R3", FormFactor: "USB-C PIN+ NFC", Branding: "Token2", Prefix: "76104"},
	{Revision: "R3", FormFactor: "FIDO Card", Branding: "Token2", Prefix: "76202"},
	{Revision: "R3", FormFactor: "FIDO Card without ISO 7816", Branding: "Unbranded", Prefix: "86106"},
	{Revision: "R3", FormFactor: "FIDO Card with ISO 7816", Branding: "Unbranded", Prefix: "76106"},

	{Revision: "R3.1", FormFactor: "USB-A PIN+ NFC", Branding: "Token2", Prefix: "76105"},
	{Revision: "R3.1", FormFactor: "USB-A PIN+ NFC", Branding: "Unbranded", Prefix: "26105"},
	{Revision: "R3.1", FormFactor: "Mini USB-C PIN+", Prefix: "72102"},
	{Revision: "R3.1", FormFactor: "Custom system access card", Branding: "Custom", SerialFrom: 70000001, SerialTo: 70002000},

	{Revision: "R3.2", FormFactor: "Dual PIN+ NFC", Branding: "Token2", Prefix: "77103"},
	{Revision: "R3.2", FormFactor: "Dual PIN+ NFC", Branding: "Unbranded", Prefix: "24103"},
	{Revision: "R3.2", FormFactor: "Mini USB-A PIN+", Prefix: "72101"},
	{Revision: "R3.2", FormFactor: "Bio3 Dual A+C PIN+", Branding: "Token2", Prefix: "72103"},
	{Revision: "R3.2", FormFactor: "Bio3 Dual A+C PIN+", Branding: "Unbranded", Prefix: "22103"},

	{Revision: "R3.3", FormFactor: "USB-A NFC PIN+ PIV+", Branding: "Token2", Prefix: "66105"},
	{Revision: "R3.3", FormFactor: "USB-C NFC PIN+ PIV+", Branding: "Token2", Prefix: "66104"},
	{Revision: "R3.3", FormFactor: "Dual NFC PIN+ PIV+", Branding: "Token2", Prefix: "66103"},
	{Revision: "R3.3", FormFactor: "USB-A NFC PIN+ PIV+", Branding: "Unbranded", Prefix: "66107"},
	{Revision: "R3.3", FormFactor: "USB-C NFC PIN+ PIV+", Branding: "Unbranded", Prefix: "66106"},
	{Revision: "R3.3", FormFactor: "Dual NFC PIN+ PIV+", Branding: "Unbranded", Prefix: "66114"},
	{Revision: "R3.3", FormFactor: "Dual NFC PIN+ PIV+", Branding: "Unbranded Octo", Prefix: "66113"},
	{Revision: "R3.3", FormFactor: "FIDO Card NFC with ISO 7816 PIN+ PIV+", Branding: "Token2", Prefix: "66202"},
	{Revision: "R3.3", FormFactor: "FIDO Card PIN+ PIV+", Branding: "Unbranded", Prefix: "66102"},
	{Revision: "R3.3", FormFactor: "FIDO Card NFC with ISO 7816 PIN+ PIV+", Branding: "Unbranded", Prefix: "66302"},
	{Revision: "R3.3", FormFactor: "Mini USB-A PIN+ PIV+", Prefix: "66101"},
	{Revision: "R3.3", FormFactor: "Mini USB-C PIN+ PIV+", Prefix: "66111"},
	{Revision: "R3.3", FormFactor: "Dual Bio3 PIN+ PIV+", Branding: "Token2", Prefix: "72113"},
	{Revision: "R3.3", FormFactor: "Dual Bio3 PIN+ PIV+", Branding: "Unbranded", Prefix: "24133"},

	{Revision: "R3.4", FormFactor: "PIN+ Dual Ace PIV+ OTP Protection", Branding: "Token2", Prefix: "65103"},
	{Revision: "R3.4", FormFactor: "Dual Bio3 PIV+ OTP Protection", Branding: "Token2", Prefix: "72114"},
	{Revision: "R3.4", FormFactor: "Mini USB-A PIN+ PIV+ OTP Protection", Prefix: "65101"},
	{Revision: "R3.4", FormFactor: "Mini USB-C PIN+ PIV+ OTP Protection", Prefix: "65111"},
}

// Models returns the built-in Token2 model catalog.
func Models() []Model {
	return slices.Clone(models)
}

// Identity contains the model information derived from a full Token2 serial
// number. Prefix, CheckDigit and Suffix are empty for range-based models.
type Identity struct {
	SerialNumber string
	Prefix       string
	CheckDigit   byte
	Suffix       string
	Model        Model
}

// Identify looks up a full Token2 serial number in the built-in model catalog.
func Identify(serialNumber string) (Identity, bool) {
	identity := Identity{SerialNumber: serialNumber}
	if len(serialNumber) < 7 {
		return identity, false
	}

	for i := range len(serialNumber) {
		if serialNumber[i] < '0' || serialNumber[i] > '9' {
			return identity, false
		}
	}

	serial, err := strconv.ParseUint(serialNumber, 10, 64)
	if err != nil {
		return identity, false
	}

	for _, model := range models {
		if model.SerialFrom != 0 && serial >= model.SerialFrom && serial <= model.SerialTo {
			identity.Model = model

			return identity, true
		}
	}

	prefix := serialNumber[:5]
	identity.Prefix = prefix
	identity.CheckDigit = serialNumber[5]
	identity.Suffix = serialNumber[6:]

	for _, model := range models {
		if model.Prefix == prefix {
			identity.Model = model

			return identity, true
		}
	}

	return identity, false
}
