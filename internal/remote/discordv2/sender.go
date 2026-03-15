package discordv2

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// MessageSender abstracts Discord message delivery so callers do not need to
// branch on bot-API vs webhook mode.
type MessageSender interface {
	// Send delivers a text message. If fileName is non-empty and fileData is
	// non-nil the message includes a file attachment.
	Send(ctx context.Context, content, fileName string, fileData []byte) error

	// SendEmbed delivers a Discord embed.
	SendEmbed(ctx context.Context, embed *discordgo.MessageEmbed) error
}

// sessionSender delivers messages through a discordgo.Session (bot-API mode).
type sessionSender struct {
	session   *discordgo.Session
	channelID string
}

func newSessionSender(session *discordgo.Session, channelID string) *sessionSender {
	return &sessionSender{session: session, channelID: channelID}
}

func (s *sessionSender) Send(_ context.Context, content, fileName string, fileData []byte) error {
	if len(fileData) > 0 && fileName != "" {
		_, err := s.session.ChannelMessageSendComplex(s.channelID, &discordgo.MessageSend{
			Content: content,
			File: &discordgo.File{
				Name:        fileName,
				ContentType: contentTypeForFile(fileName),
				Reader:      newByteReader(fileData),
			},
		})
		return err
	}
	_, err := s.session.ChannelMessageSend(s.channelID, content)
	return err
}

func (s *sessionSender) SendEmbed(_ context.Context, embed *discordgo.MessageEmbed) error {
	_, err := s.session.ChannelMessageSendEmbed(s.channelID, embed)
	return err
}
