package tg

type Message interface {
	Where() Peer
	Sender() Peer
	Author() Peer
	IsForwarded() bool
	Content() string
}

type ChatMessageData struct {
	Peer        Peer
	FromPeer    Peer
	FwdFromPeer Peer
	Content     string
}

type ChatMessage struct {
	ChatMessageData
}

func (c *ChatMessage) Where() Peer {
	return c.ChatMessageData.Peer
}

func (c *ChatMessage) Sender() Peer {
	return c.FromPeer
}

func (c *ChatMessage) Author() Peer {
	if c.IsForwarded() {
		return c.FwdFromPeer
	}

	return c.FromPeer
}

func (c *ChatMessage) Content() string {
	return c.ChatMessageData.Content
}

func (c *ChatMessage) IsForwarded() bool {
	return c.FwdFromPeer != nil
}
