---
name: discord-bot
description: Patterns pour créer un bot Discord en Go avec discordgo. Utiliser pour les commandes slash, webhooks, alertes.
allowed-tools: Read, Grep, Glob, Edit, Write, Bash
---

# Discord Bot Patterns (Go + discordgo)

## Setup de base

```go
package discord

import "github.com/bwmarrin/discordgo"

type Bot struct {
    session *discordgo.Session
    guildID string
}

func NewBot(token string) (*Bot, error) {
    s, err := discordgo.New("Bot " + token)
    if err != nil {
        return nil, err
    }
    return &Bot{session: s}, nil
}
```

## Slash Commands

```go
var commands = []*discordgo.ApplicationCommand{
    {
        Name:        "top",
        Description: "Affiche les produits les plus rentables",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "period",
                Description: "Période (7d, 30d)",
                Required:    false,
            },
        },
    },
}

func (b *Bot) handleTop(s *discordgo.Session, i *discordgo.InteractionCreate) {
    // Récupérer les données
    products := b.analyzer.GetTopProducts(7)

    // Formater le message
    embed := &discordgo.MessageEmbed{
        Title: "🏆 Top Produits Rentables",
        Fields: formatProducts(products),
    }

    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Embeds: []*discordgo.MessageEmbed{embed},
        },
    })
}
```

## Webhooks (alertes)

```go
func SendAlert(webhookURL string, alert Alert) error {
    embed := &discordgo.MessageEmbed{
        Title:       "🚨 " + alert.Title,
        Description: alert.Description,
        Color:       0xFF0000,
        Fields: []*discordgo.MessageEmbedField{
            {Name: "Prix", Value: fmt.Sprintf("%.2f€", alert.Price)},
            {Name: "ROI", Value: fmt.Sprintf("+%.0f%%", alert.ROI)},
        },
        URL: alert.ProductURL,
    }

    _, err := discordgo.New("").WebhookExecute(webhookID, webhookToken, true,
        &discordgo.WebhookParams{Embeds: []*discordgo.MessageEmbed{embed}})
    return err
}
```
