package ambiguity

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrOptionNotFound       = errors.New("option is not defined")
	ErrOptionUnexpectedType = errors.New("option defined with unexpected type")
)

// Helper class to obtain values from ApplicationCommandInteractionDataOption
type CommandOptions struct {
	session *discordgo.Session
	options []*discordgo.ApplicationCommandInteractionDataOption
}

func (c *CommandOptions) Option(name string) (*discordgo.ApplicationCommandInteractionDataOption, error) {
	for _, option := range c.options {
		if option.Name == name {
			return option, nil
		}
	}

	return nil, fmt.Errorf("%q: %w", name, ErrOptionNotFound)
}

func (c *CommandOptions) String(name string) (string, error) {
	option, err := c.Option(name)
	if err != nil {
		return "", nil
	}

	if err := c.expect(option, discordgo.ApplicationCommandOptionString); err != nil {
		return "", err
	}

	return option.StringValue(), nil
}

func (c *CommandOptions) StringDefault(name string, def string) string {
	value, err := c.String(name)
	if err != nil {
		return def
	}

	return value
}

func (c *CommandOptions) Integer(name string) (int64, error) {
	option, err := c.Option(name)
	if err != nil {
		return 0, nil
	}

	if err := c.expect(option, discordgo.ApplicationCommandOptionInteger); err != nil {
		return 0, err
	}

	return option.IntValue(), nil
}

func (c *CommandOptions) IntegerDefault(name string, def int64) int64 {
	value, err := c.Integer(name)
	if err != nil {
		return def
	}

	return value
}

func (c *CommandOptions) Float(name string) (float64, error) {
	option, err := c.Option(name)
	if err != nil {
		return 0, nil
	}

	if err := c.expect(option, discordgo.ApplicationCommandOptionNumber); err != nil {
		return 0, err
	}

	return option.FloatValue(), nil
}

func (c *CommandOptions) FloatDefault(name string, def float64) float64 {
	value, err := c.Float(name)
	if err != nil {
		return def
	}

	return value
}

func (c *CommandOptions) Bool(name string) (bool, error) {
	option, err := c.Option(name)
	if err != nil {
		return false, nil
	}

	if err := c.expect(option, discordgo.ApplicationCommandOptionBoolean); err != nil {
		return false, err
	}

	return option.BoolValue(), nil
}

func (c *CommandOptions) BoolDefault(name string, def bool) bool {
	value, err := c.Bool(name)
	if err != nil {
		return def
	}

	return value
}

func (c *CommandOptions) User(name string) (*discordgo.User, error) {
	option, err := c.Option(name)
	if err != nil {
		return nil, nil
	}

	if err := c.expect(option, discordgo.ApplicationCommandOptionUser); err != nil {
		return nil, err
	}

	return option.UserValue(c.session), nil
}

func (c *CommandOptions) Role(name string, guildID string) (*discordgo.Role, error) {
	option, err := c.Option(name)
	if err != nil {
		return nil, nil
	}

	if err := c.expect(option, discordgo.ApplicationCommandOptionRole); err != nil {
		return nil, err
	}

	return option.RoleValue(c.session, guildID), nil
}

func (c *CommandOptions) Channel(name string) (*discordgo.Channel, error) {
	option, err := c.Option(name)
	if err != nil {
		return nil, nil
	}

	if err := c.expect(option, discordgo.ApplicationCommandOptionChannel); err != nil {
		return nil, err
	}

	return option.ChannelValue(c.session), nil
}

func (c *CommandOptions) Subcommand(name string) (*CommandOptions, error) {
	option, err := c.Option(name)
	if err != nil {
		return nil, nil
	}

	if err := c.expect(option, discordgo.ApplicationCommandOptionSubCommand); err != nil {
		return nil, err
	}

	return &CommandOptions{
		session: c.session,
		options: option.Options,
	}, nil
}

func (c *CommandOptions) SubcommandGroup(name string) (*CommandOptions, error) {
	option, err := c.Option(name)
	if err != nil {
		return nil, nil
	}

	if err := c.expect(option, discordgo.ApplicationCommandOptionSubCommandGroup); err != nil {
		return nil, err
	}

	return &CommandOptions{
		session: c.session,
		options: option.Options,
	}, nil
}

func (c *CommandOptions) expect(
	option *discordgo.ApplicationCommandInteractionDataOption,
	expected discordgo.ApplicationCommandOptionType,
) error {
	if option.Type != expected {
		return fmt.Errorf("expected %q found %q: %w", expected, option.Type, ErrOptionNotFound)
	}

	return nil
}
