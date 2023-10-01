package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func generateRandomData(channel chan []byte) {
	for p := range channel {
		randomData := make([]byte, 1004) // Generate random data of 1 KB for testing
		rand.Read(randomData)
		channel <- p
		channel <- randomData
	}
}

func handleVoice(channel chan *discordgo.Packet) {
	http.HandleFunc("/audio", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "audio/ogg")
		w.Header().Set("Transfer-Encoding", "chunked")

		//save the audio and "append" to a buffer

		for {
			// randomData := make([]byte, 1024) // Generate random data of 1 KB for testing
			// rand.Read(randomData)
			// channel <- randomData
			// time.Sleep(time.Second)

			select {
			case data, ok := <-channel:
				if !ok {
					fmt.Println("Closed")
					return
				}

				w.Write(data.Opus)

				// Flush the data to the client immediately
				w.(http.Flusher).Flush()

				fmt.Println("Received:", data)

			}
		}
	})
	port := ":8080"
	fmt.Printf("Server listening on port %s\n", port)
	http.ListenAndServe(port, nil)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		Token     = os.Getenv("BOT_TOKEN")
		ChannelID = os.Getenv("CHANNEL_ID")
		ServerID  = os.Getenv("SERVER_ID")
	)

	discord, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session:", err)
		return
	}

	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	call, err := discord.ChannelVoiceJoin(ServerID, ChannelID, true, false)
	if err != nil {
		fmt.Println("failed to join voice channel:", err)
		return
	}

	go func() {
		time.Sleep(15 * time.Second)
		close(call.OpusRecv)
		call.Close()
	}()

	handleVoice(call.OpusRecv)

	//Exit Bot
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}
