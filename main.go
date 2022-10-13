package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	env := os.Getenv("BOT_ENV")

	if err != nil && env == "dev" {
		log.Fatal("Error loading .env file")
	}

	discordToken := os.Getenv("DISCORD_BOT_TOKEN")
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", discordToken))
	if err != nil {
		log.Fatal(err)
	}

	discord.AddHandler(replyMessage)
	discord.Identify.Intents = discordgo.IntentsGuildMessages

	fmt.Println("Bot is starting.")

	err = discord.Open()
	if err != nil {
		fmt.Println(err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	fmt.Println("Disconnecting...")
	discord.Close()
	fmt.Println("Disconnected...")
}

func replyMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.Contains(m.Content, "/resultado") {
		fmt.Println(fmt.Sprintf(`Received message: %s`, m.Content))
		messageWords := strings.Split(m.Content, " ")

		if len(messageWords) != 3 {
			s.ChannelMessageSendReply(
				m.ChannelID,
				`Meu patrão, o formato do comando deve ser: "/resultado nomedojogo numerodoconcurso"`,
				m.MessageReference,
			)
			return
		}

		gameName := messageWords[1]
		contestNumber := messageWords[2]

		_, ok := getPossibleGames()[gameName]

		if !ok {
			s.ChannelMessageSendReply(
				m.ChannelID,
				`Minha joia, o jogo informado é inválido. Valores de jogos possíveis: megasena, quina, lotofacil, duplasena, timemania`,
				m.MessageReference,
			)
			return
		}

		s.ChannelMessageSendReply(
			m.ChannelID,
			`Opa meu consagrado, deixa eu conferir aqui os resultados`,
			m.MessageReference,
		)

		gameResultNumbers := getGameResult(gameName, contestNumber)
		replyMessage := fmt.Sprintf("Resultado da %s: %s", gameName, strings.Join(gameResultNumbers, ", "))
		s.ChannelMessageSendReply(m.ChannelID, replyMessage, m.MessageReference)
	}
}

func getGameResult(gameName string, contestNumber string) []string {
	page := rod.
		New().
		MustConnect().
		MustPage(fmt.Sprintf("http://loterias.caixa.gov.br/wps/portal/loterias/landing/%s/", gameName)).
		MustWindowMaximize()

	page.
		MustElement("#buscaConcurso").
		Timeout(time.Second * 10).
		MustFocus().
		MustInput(contestNumber).
		MustPress(input.Enter)

	time.Sleep(time.Second * 2)

	resultNumberRows := page.MustElements(fmt.Sprintf(".%s > li", gameName))

	var resultNumbers []string
	for _, row := range resultNumberRows {
		resultNumbers = append(resultNumbers, row.MustText())
	}

	return resultNumbers
}

func getPossibleGames() map[string]struct{} {
	return map[string]struct{}{
		"megasena":  {},
		"quina":     {},
		"lotofacil": {},
		"lotomania": {},
		"duplasena": {},
		"timemania": {},
	}
}
