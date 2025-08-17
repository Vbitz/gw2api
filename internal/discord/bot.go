package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"j5.nz/gw2/internal/gw2api"
)

// Bot represents the Discord bot
type Bot struct {
	session *discordgo.Session
	client  *gw2api.Client
}

// NewBot creates a new Discord bot instance
func NewBot(token string, gw2Client *gw2api.Client) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	bot := &Bot{
		session: session,
		client:  gw2Client,
	}

	// Register slash command handlers
	session.AddHandler(bot.handleInteraction)
	session.AddHandler(bot.ready)

	return bot, nil
}

// Start starts the Discord bot
func (b *Bot) Start() error {
	return b.session.Open()
}

// Stop stops the Discord bot
func (b *Bot) Stop() error {
	return b.session.Close()
}

// ready handles the ready event and registers slash commands
func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("Discord bot logged in as: %v#%v\n", event.User.Username, event.User.Discriminator)

	// Register slash commands
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "gw2-build",
			Description: "Get the current Guild Wars 2 build information",
		},
		{
			Name:        "gw2-achievement",
			Description: "Get information about a specific achievement",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Achievement ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "gw2-currency",
			Description: "Get information about a specific currency",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Currency ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "gw2-item",
			Description: "Get information about a specific item",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Item ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "gw2-world",
			Description: "Get information about a specific world/server",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "World ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "gw2-skill",
			Description: "Get information about a specific skill",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "id",
					Description: "Skill ID",
					Required:    true,
				},
			},
		},
		{
			Name:        "gw2-prices",
			Description: "Get trading post prices for an item",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "item_id",
					Description: "Item ID to check prices for",
					Required:    true,
				},
			},
		},
	}

	// Register commands
	for _, command := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", command)
		if err != nil {
			fmt.Printf("Cannot create slash command %q: %v\n", command.Name, err)
		}
	}
}

// handleInteraction handles slash command interactions
func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "" {
		return
	}

	var response string
	var err error

	switch i.ApplicationCommandData().Name {
	case "gw2-build":
		response, err = b.handleBuildCommand()
	case "gw2-achievement":
		id := int(i.ApplicationCommandData().Options[0].IntValue())
		response, err = b.handleAchievementCommand(id)
	case "gw2-currency":
		id := int(i.ApplicationCommandData().Options[0].IntValue())
		response, err = b.handleCurrencyCommand(id)
	case "gw2-item":
		id := int(i.ApplicationCommandData().Options[0].IntValue())
		response, err = b.handleItemCommand(id)
	case "gw2-world":
		id := int(i.ApplicationCommandData().Options[0].IntValue())
		response, err = b.handleWorldCommand(id)
	case "gw2-skill":
		id := int(i.ApplicationCommandData().Options[0].IntValue())
		response, err = b.handleSkillCommand(id)
	case "gw2-prices":
		id := int(i.ApplicationCommandData().Options[0].IntValue())
		response, err = b.handlePricesCommand(id)
	default:
		response = "Unknown command"
	}

	if err != nil {
		response = fmt.Sprintf("Error: %v", err)
	}

	// Respond to the interaction
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})
	if err != nil {
		fmt.Printf("Error responding to interaction: %v\n", err)
	}
}

// Command handlers
func (b *Bot) handleBuildCommand() (string, error) {
	ctx := context.Background()
	build, err := b.client.GetBuild(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("**Current GW2 Build**: %d", build.ID), nil
}

func (b *Bot) handleAchievementCommand(id int) (string, error) {
	ctx := context.Background()
	achievement, err := b.client.GetAchievement(ctx, id)
	if err != nil {
		return "", err
	}

	response := fmt.Sprintf("**%s** (ID: %d)\n", achievement.Name, achievement.ID)
	response += fmt.Sprintf("*%s*\n", achievement.Description)
	if len(achievement.Tiers) > 0 {
		response += fmt.Sprintf("**Tiers**: %d\n", len(achievement.Tiers))
	}
	if len(achievement.Flags) > 0 {
		response += fmt.Sprintf("**Flags**: %s", strings.Join(achievement.Flags, ", "))
	}

	return response, nil
}

func (b *Bot) handleCurrencyCommand(id int) (string, error) {
	ctx := context.Background()
	currency, err := b.client.GetCurrency(ctx, id)
	if err != nil {
		return "", err
	}

	response := fmt.Sprintf("**%s** (ID: %d)\n", currency.Name, currency.ID)
	response += fmt.Sprintf("*%s*\n", currency.Description)
	response += fmt.Sprintf("**Order**: %d", currency.Order)

	return response, nil
}

func (b *Bot) handleItemCommand(id int) (string, error) {
	ctx := context.Background()
	item, err := b.client.GetItem(ctx, id)
	if err != nil {
		return "", err
	}

	response := fmt.Sprintf("**%s** (ID: %d)\n", item.Name, item.ID)
	response += fmt.Sprintf("*%s*\n", item.Description)
	response += fmt.Sprintf("**Type**: %s | **Rarity**: %s | **Level**: %d\n", item.Type, item.Rarity, item.Level)
	if item.VendorValue > 0 {
		response += fmt.Sprintf("**Vendor Value**: %d copper", item.VendorValue)
	}

	return response, nil
}

func (b *Bot) handleWorldCommand(id int) (string, error) {
	ctx := context.Background()
	world, err := b.client.GetWorld(ctx, id)
	if err != nil {
		return "", err
	}

	response := fmt.Sprintf("**%s** (ID: %d)\n", world.Name, world.ID)
	response += fmt.Sprintf("**Population**: %s", world.Population)

	return response, nil
}

func (b *Bot) handleSkillCommand(id int) (string, error) {
	ctx := context.Background()
	skill, err := b.client.GetSkill(ctx, id)
	if err != nil {
		return "", err
	}

	response := fmt.Sprintf("**%s** (ID: %d)\n", skill.Name, skill.ID)
	response += fmt.Sprintf("*%s*\n", skill.Description)
	if skill.Type != "" {
		response += fmt.Sprintf("**Type**: %s\n", skill.Type)
	}
	if skill.WeaponType != "" {
		response += fmt.Sprintf("**Weapon Type**: %s", skill.WeaponType)
	}

	return response, nil
}

func (b *Bot) handlePricesCommand(id int) (string, error) {
	ctx := context.Background()
	price, err := b.client.GetCommercePrice(ctx, id)
	if err != nil {
		return "", err
	}

	response := fmt.Sprintf("**Trading Post Prices for Item %d**\n", id)
	if price.Buys.Quantity > 0 {
		response += fmt.Sprintf("**Buy Orders**: %d @ %d copper each\n", price.Buys.Quantity, price.Buys.UnitPrice)
	} else {
		response += "**Buy Orders**: None\n"
	}

	if price.Sells.Quantity > 0 {
		response += fmt.Sprintf("**Sell Listings**: %d @ %d copper each", price.Sells.Quantity, price.Sells.UnitPrice)
	} else {
		response += "**Sell Listings**: None"
	}

	return response, nil
}
