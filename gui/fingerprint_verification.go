package gui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
	"github.com/twstrike/coyim/client"
	"github.com/twstrike/coyim/config"
	"github.com/twstrike/coyim/i18n"
	"github.com/twstrike/otr3"
)

func buildVerifyFingerprintDialog(accountName string, ourFp []byte, uid string, theirFp []byte) *gtk.Dialog {
	var message string
	var builderName string

	if theirFp == nil {
		builderName = "VerifyFingerprintUnknown"
		message = fmt.Sprintf(i18n.Local(
			"You can't verify the fingerprint for %s yet.\n"+
				"You first have to start an encrypted conversation with them.",
		), uid)

	} else {
		m := i18n.Local(`
Is this the correct fingerprint for %[1]s?

Fingerprint for you (%[3]s):
  %[4]s

Purported fingerprint for %[1]s:
  %[2]s
	`)

		message = fmt.Sprintf(m,
			uid,
			config.FormatFingerprint(theirFp),
			accountName,
			config.FormatFingerprint(ourFp),
		)

		builderName = "VerifyFingerprint"
	}

	builder := builderForDefinition(builderName)

	obj, _ := builder.GetObject("dialog")
	dialog := obj.(*gtk.Dialog)

	obj, _ = builder.GetObject("message")
	l := obj.(*gtk.Label)
	l.SetText(message)

	dialog.SetTitle(fmt.Sprintf(i18n.Local("Verify fingerprint for %s"), uid))
	return dialog
}

func getFingerprintsFor(conversation *otr3.Conversation) ([]byte, []byte) {
	var ourFp, theirFp []byte
	ourKey := conversation.GetOurCurrentKey()
	if ourKey != nil {
		ourFp = ourKey.PublicKey().Fingerprint()
	}

	theirKey := conversation.GetTheirKey()
	if theirKey != nil {
		theirFp = theirKey.Fingerprint()
	}

	return ourFp, theirFp
}

func verifyFingerprintDialog(account *account, uid string, parent *gtk.Window) {
	accountConfig := account.session.CurrentAccount
	conversation := account.session.GetConversationWith(uid)
	ourFp, theirFp := getFingerprintsFor(conversation)

	dialog := buildVerifyFingerprintDialog(accountConfig.Account, ourFp, uid, theirFp)
	defer dialog.Destroy()

	dialog.SetTransientFor(parent)
	dialog.ShowAll()

	responseType := gtk.ResponseType(dialog.Run())
	switch responseType {
	case gtk.RESPONSE_YES:
		account.ExecuteCmd(client.AuthorizeFingerprintCmd{
			Account:     accountConfig,
			Peer:        uid,
			Fingerprint: theirFp,
		})
	}
}
