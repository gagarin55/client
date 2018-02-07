// Copyright 2016 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package client

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/protocol/chat1"
)

type ChatUI struct {
	libkb.Contextified
	terminal            libkb.TerminalUI
	noOutput            bool
	lastPercentReported int
	searchMutex         sync.Mutex
}

func (c *ChatUI) ChatAttachmentUploadOutboxID(context.Context, chat1.ChatAttachmentUploadOutboxIDArg) error {
	return nil
}

func (c *ChatUI) ChatAttachmentUploadStart(context.Context, chat1.ChatAttachmentUploadStartArg) error {
	if c.noOutput {
		return nil
	}
	w := c.terminal.ErrorWriter()
	fmt.Fprintf(w, "Attachment upload "+ColorString(c.G(), "green", "starting")+"\n")
	return nil
}

func (c *ChatUI) ChatAttachmentUploadProgress(ctx context.Context, arg chat1.ChatAttachmentUploadProgressArg) error {
	if c.noOutput {
		return nil
	}
	percent := int((100 * arg.BytesComplete) / arg.BytesTotal)
	if c.lastPercentReported == 0 || percent == 100 || percent-c.lastPercentReported >= 10 {
		w := c.terminal.ErrorWriter()
		fmt.Fprintf(w, "Attachment upload progress %d%% (%d of %d bytes uploaded)\n", percent, arg.BytesComplete, arg.BytesTotal)
		c.lastPercentReported = percent
	}
	return nil
}

func (c *ChatUI) ChatAttachmentUploadDone(context.Context, int) error {
	if c.noOutput {
		return nil
	}
	w := c.terminal.ErrorWriter()
	fmt.Fprintf(w, "Attachment upload "+ColorString(c.G(), "magenta", "finished")+"\n")
	return nil
}

func (c *ChatUI) ChatAttachmentPreviewUploadStart(context.Context, chat1.ChatAttachmentPreviewUploadStartArg) error {
	if c.noOutput {
		return nil
	}
	w := c.terminal.ErrorWriter()
	fmt.Fprintf(w, "Attachment preview upload "+ColorString(c.G(), "green", "starting")+"\n")
	return nil
}

func (c *ChatUI) ChatAttachmentPreviewUploadDone(context.Context, int) error {
	if c.noOutput {
		return nil
	}
	w := c.terminal.ErrorWriter()
	fmt.Fprintf(w, "Attachment preview upload "+ColorString(c.G(), "magenta", "finished")+"\n")
	return nil
}

func (c *ChatUI) ChatAttachmentDownloadStart(context.Context, int) error {
	if c.noOutput {
		return nil
	}
	w := c.terminal.ErrorWriter()
	fmt.Fprintf(w, "Attachment download "+ColorString(c.G(), "green", "starting")+"\n")
	return nil
}

func (c *ChatUI) ChatAttachmentDownloadProgress(ctx context.Context, arg chat1.ChatAttachmentDownloadProgressArg) error {
	if c.noOutput {
		return nil
	}
	percent := int((100 * arg.BytesComplete) / arg.BytesTotal)
	if c.lastPercentReported == 0 || percent == 100 || percent-c.lastPercentReported >= 10 {
		w := c.terminal.ErrorWriter()
		fmt.Fprintf(w, "Attachment download progress %d%% (%d of %d bytes downloaded)\n", percent, arg.BytesComplete, arg.BytesTotal)
		c.lastPercentReported = percent
	}
	return nil
}

func (c *ChatUI) ChatAttachmentDownloadDone(context.Context, int) error {
	if c.noOutput {
		return nil
	}
	w := c.terminal.ErrorWriter()
	fmt.Fprintf(w, "Attachment download "+ColorString(c.G(), "magenta", "finished")+"\n")
	return nil
}

func (c *ChatUI) ChatInboxConversation(ctx context.Context, arg chat1.ChatInboxConversationArg) error {
	return nil
}

func (c *ChatUI) ChatInboxFailed(ctx context.Context, arg chat1.ChatInboxFailedArg) error {
	return nil
}

func (c *ChatUI) ChatInboxUnverified(ctx context.Context, arg chat1.ChatInboxUnverifiedArg) error {
	return nil
}

func (c *ChatUI) ChatThreadCached(ctx context.Context, arg chat1.ChatThreadCachedArg) error {
	return nil
}

func (c *ChatUI) ChatThreadFull(ctx context.Context, arg chat1.ChatThreadFullArg) error {
	return nil
}

func (c *ChatUI) ChatConfirmChannelDelete(ctx context.Context, arg chat1.ChatConfirmChannelDeleteArg) (bool, error) {
	term := c.G().UI.GetTerminalUI()
	term.Printf("WARNING: This will destroy this chat channel and remove it from all members' inbox\n\n")
	confirm := fmt.Sprintf("nuke %s", arg.Channel)
	response, err := term.Prompt(PromptDescriptorDeleteRootTeam,
		fmt.Sprintf("** if you are sure, please type: %q > ", confirm))
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(response) == confirm, nil
}

func (c *ChatUI) ChatSearchHit(ctx context.Context, arg chat1.ChatSearchHitArg) error {
	// We don't want results written out of order
	c.searchMutex.Lock()
	defer c.searchMutex.Unlock()
	if c.noOutput {
		return nil
	}
	const maxContext = 100 // Only show the first 100 chars of the context messages
	getContext := func(uiMsg chat1.UIMessage) string {
		if uiMsg.IsValid() && uiMsg.GetMessageType() == chat1.MessageType_TEXT {
			msgBody := uiMsg.Valid().MessageBody.Text().Body
			sliceIdx := len(msgBody) - maxContext
			if sliceIdx < 0 {
				sliceIdx = 0
			}
			return msgBody[sliceIdx:] + "\n"
		}
		return ""
	}

	highlightHits := func(uiMsg chat1.UIMessage, hits []string) string {
		if uiMsg.IsValid() && uiMsg.GetMessageType() == chat1.MessageType_TEXT {
			msgBody := uiMsg.Valid().MessageBody.Text().Body
			var hitText string
			for _, hit := range hits {
				hitText = strings.Replace(msgBody, hit, ColorString(c.G(), "red", hit), -1)
			}

			return hitText
		}
		return ""
	}
	hitText := highlightHits(arg.HitMessage, arg.Hits)
	if hitText != "" {
		w := c.terminal.ErrorWriter()
		fmt.Fprintf(w, getContext(arg.PrevMessage))
		fmt.Fprintln(w, hitText)
		fmt.Fprintf(w, getContext(arg.NextMessage))
		fmt.Fprintln(w, "")
	}
	return nil
}

func (c *ChatUI) ChatSearchDone(ctx context.Context, arg chat1.ChatSearchDoneArg) error {
	if c.noOutput {
		return nil
	}
	w := c.terminal.ErrorWriter()
	fmt.Fprintf(w, "Search complete. Found %d results.", arg.NumHits)
	fmt.Fprintln(w, "")
	return nil
}
