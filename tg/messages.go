package tg

type Message interface {
	ID() int64
	Where() Peer
	Sender() Peer
	Author() Peer
	IsForwarded() bool
	Content() string
	IsOutgoing() bool
}

type ChatMessageData struct {
	ID          int64
	Peer        Peer
	FromPeer    Peer
	FwdFromPeer Peer
	Content     string
	IsOutgoing  bool
}

type ChatMessage struct {
	ChatMessageData
}

func (c *ChatMessage) ID() int64 {
	return c.ChatMessageData.ID
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

func (c *ChatMessage) IsOutgoing() bool {
	return c.ChatMessageData.IsOutgoing
}
