# Message Formatting / Форматирование сообщений

This document describes the message formatting feature for the gotg library.

## Overview / Обзор

The message formatting feature allows you to create rich text messages using a tree-based structure. Instead of plain text strings, you can compose messages from formatted chunks with various styles.

Функция форматирования сообщений позволяет создавать сообщения с богатым текстом, используя древовидную структуру. Вместо простых текстовых строк вы можете составлять сообщения из форматированных кусков с различными стилями.

## Core Concepts / Основные концепции

### MessageChunk Interface

All formatting elements implement the `MessageChunk` interface:

```go
type MessageChunk interface {
    ToStyledTextOptions() []styling.StyledTextOption
}
```

### Chunk Types / Типы кусков

1. **TextChunk** - Plain text without formatting / Простой текст без форматирования
2. **StyledTextChunk** - Text with styling / Текст со стилем
3. **ContainerChunk** - Container for multiple chunks / Контейнер для нескольких кусков
4. **LinkChunk** - Text with URL / Текст со ссылкой
5. **PreCodeChunk** - Code block with syntax highlighting / Блок кода с подсветкой синтаксиса

## API Methods / Методы API

### Sending Messages / Отправка сообщений

```go
// Send formatted message
peer.SendMessageFormatted(ctx, chunk)

// Reply with formatted message
message.ReplyFormatted(ctx, tg, chunk)
messageRef.ReplyFormatted(ctx, tg, chunk)
```

### Helper Functions / Вспомогательные функции

```go
// Basic text
tg.Text("Hello world")

// Styled text
tg.Bold("bold text")
tg.Italic("italic text")
tg.Code("inline code")
tg.Strike("strikethrough")
tg.Underline("underlined")
tg.Spoiler("spoiler text")

// Links and code blocks
tg.Link("link text", "https://example.com")
tg.PreCode("func main() {}", "go")

// Containers
tg.Container(
    tg.Text("Hello "),
    tg.Bold("world"),
    tg.Text("!"),
)
```

## Examples / Примеры

### Basic Usage / Базовое использование

```go
chunk := tg.Container(
    tg.Text("This is "),
    tg.Bold("bold"),
    tg.Text(" and "),
    tg.Italic("italic"),
    tg.Text(" text."),
)

_, err := peer.SendMessageFormatted(ctx, chunk)
```

### Complex Formatting / Сложное форматирование

```go
message := tg.Container(
    tg.Bold("Important Notice!"),
    tg.Text("\n\nPlease visit our "),
    tg.Link("website", "https://example.com"),
    tg.Text(" for more information about "),
    tg.Spoiler("secret features"),
    tg.Text(".\n\nCode example:\n"),
    tg.PreCode("package main\n\nfunc main() {\n    fmt.Println(\"Hello!\")\n}", "go"),
)

_, err := peer.SendMessageFormatted(ctx, message)
```

### Nested Containers / Вложенные контейнеры

```go
nested := tg.Container(
    tg.Text("This is a "),
    tg.Container(
        tg.Bold("nested"),
        tg.Text(" "),
        tg.Italic("container"),
    ),
    tg.Text(" example."),
)

_, err := message.ReplyFormatted(ctx, tg, nested)
```

## Russian Language Support / Поддержка русского языка

The formatting system fully supports Russian and other Unicode text:

```go
russianMessage := tg.Container(
    tg.Bold("Важное уведомление!"),
    tg.Text("\n\nПожалуйста, посетите наш "),
    tg.Link("сайт", "https://example.com"),
    tg.Text(" для получения дополнительной информации."),
)

_, err := peer.SendMessageFormatted(ctx, russianMessage)
```

## Style Types / Типы стилей

- `StyleBold` - **Bold text** / **Жирный текст**
- `StyleItalic` - *Italic text* / *Курсивный текст*  
- `StyleCode` - `Inline code` / `Встроенный код`
- `StylePre` - Code block / Блок кода
- `StyleStrike` - ~~Strikethrough~~ / ~~Зачёркнутый~~
- `StyleUnderline` - Underlined / Подчёркнутый
- `StyleSpoiler` - Spoiler text / Спойлер

## Examples Directory / Папка с примерами

Check out the examples in `tg/adapters/gotd/examples/`:

- `formatting_demo/` - Basic formatting demonstration
- `formatting_validation/` - Validation tests
- `russian_demo/` - Russian language examples
- `user_print_messages/` - Updated to use formatted replies

## Implementation Details / Детали реализации

The formatting system converts the chunk tree into gotd `StyledTextOption` slices, which are then passed to Telegram's message styling system. This ensures compatibility with all Telegram formatting features while providing a clean, composable API.

Система форматирования преобразует дерево кусков в срезы `StyledTextOption` библиотеки gotd, которые затем передаются в систему стилизации сообщений Telegram. Это обеспечивает совместимость со всеми функциями форматирования Telegram при предоставлении чистого, композиционного API.