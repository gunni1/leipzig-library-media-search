# User Story
As a User i want to be able to subscribe for a notification for a currently not available title on my watchlist.

# Decisions

## Telegram as notification channel
- easy to implement with a simple bot
- no further accounts needed 
- technical less complicated than sending an email

## Notifier as a separate executable
??



# Use Cases

## Subscribe for notification 

```mermaid
sequenceDiagram
    actor User
    participant LLMS as LLMS
    participant Telegram as Telegram
    participant Storage as Storage
    User->>LLMS: subscribe for notification
    activate LLMS
    LLMS->>LLMS: create link
    deactivate LLMS
    User->>Telegram: Open t.me link with title data
    Telegram-->>LLMS: Create subscription for user (chatId) 
    LLMS->>Storage: Save subscription
```

## Scheduled check subscriptions 
```mermaid
sequenceDiagram
    actor User
    participant LLMS as LLMS
      participant Telegram as Telegram
      participant Storage as Storage
      participant Library as Katalog Bibliothek
      LLMS->>Storage: Load all subscriptions
      Storage-->>LLMS: return
      LLMS->>Library: Request availability for titles
      Library-->>LLMS: return
      LLMS->>Telegram: send notification when available
      Telegram->>User: notify
```
