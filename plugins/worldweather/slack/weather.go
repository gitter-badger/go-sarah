package slack

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/oklahomer/go-sarah"
	"github.com/oklahomer/go-sarah/plugins/worldweather"
	"github.com/oklahomer/go-sarah/slack"
	"github.com/oklahomer/go-sarah/slack/webapi"
	"golang.org/x/net/context"
	"regexp"
)

var (
	identifier = "weather"
)

type pluginConfig struct {
	APIKey string `yaml:"api_key"`
}

func weather(strippedMessage string, input sarah.BotInput, config sarah.CommandConfig) (*sarah.PluginResponse, error) {
	// Share client instance with later execution
	conf, _ := config.(*pluginConfig)
	client := worldweather.NewClient(worldweather.NewConfig(conf.APIKey))
	resp, err := client.LocalWeather(context.TODO(), strippedMessage)
	// TODO err check
	if err != nil {
		logrus.Errorf("Error on weather api reqeust: %s.", err.Error())
		return slack.NewStringPluginResponse("Something went wrong with weather api request."), nil
	}

	request := resp.Data.Request[0]
	currentCondition := resp.Data.CurrentCondition[0]
	forecast := resp.Data.Weather[0]
	astronomy := forecast.Astronomy[0]
	currentDesc := fmt.Sprintf("Current weather at %s is... %s.", request.Query, currentCondition.Description[0].Content)
	primaryLabelColor := "#32CD32"   // lime green
	secondaryLabelColor := "#006400" // dark green

	attachments := []*webapi.MessageAttachment{
		// Current condition and overall description
		{
			Fallback: currentDesc,
			Pretext:  "Current Condition",
			Title:    currentDesc,
			Color:    primaryLabelColor,
			ImageURL: currentCondition.WeatherIcon[0].URL,
		},

		// Temperature
		{
			Fallback: fmt.Sprintf("Temperature: %s degrees Celsius.", currentCondition.Temperature),
			Title:    "Temperature",
			Color:    primaryLabelColor,
			Fields: []*webapi.AttachmentField{
				{
					Title: "Celsius",
					Value: string(currentCondition.Temperature),
					Short: true,
				},
			},
		},

		// Wind speed
		{
			Fallback: fmt.Sprintf("Wind speed: %s Km/h", currentCondition.WindSpeed),
			Title:    "Wind Speed",
			Color:    primaryLabelColor,
			Fields: []*webapi.AttachmentField{
				{
					Title: "kn/h",
					Value: string(currentCondition.WindSpeed),
					Short: true,
				},
			},
		},

		// Astronomy
		{
			Fallback: fmt.Sprintf("Sunrise at %s. Sunset at %s.", astronomy.Sunrise, astronomy.Sunset),
			Pretext:  "Astronomy",
			Title:    "",
			Color:    secondaryLabelColor,
			Fields: []*webapi.AttachmentField{
				{
					Title: "Sunrise",
					Value: astronomy.Sunrise,
					Short: true,
				},
				{
					Title: "Sunset",
					Value: astronomy.Sunset,
					Short: true,
				},
				{
					Title: "Moonrise",
					Value: astronomy.MoonRise,
					Short: true,
				},
				{
					Title: "Moonset",
					Value: astronomy.MoonSet,
					Short: true,
				},
			},
		},
	}

	return slack.NewPostMessagePluginResponse(input, "", attachments), nil
}

func init() {
	builder := sarah.NewCommandBuilder().
		Identifier(identifier).
		ConfigStruct(&pluginConfig{}).
		MatchPattern(regexp.MustCompile(`^\.weather`)).
		Func(weather).
		Example(".echo knock knock")
	sarah.AppendCommandBuilder(slack.SLACK, builder)
}
