package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

const ProcessedTweets = "processed_tweets"

var client *twitter.Client

func initTwitterClient() {
	config := oauth1.NewConfig(TWITTER_API_KEY, TWITTER_API_SECRET)
	token := oauth1.NewToken(TWITTER_ACCESS_TOKEN_KEY, TWITTER_ACCESS_TOKEN_SECRET)
	httpClient := config.Client(oauth1.NoContext, token)
	client = twitter.NewClient(httpClient)
}

func processTweet(tweet twitter.Tweet) {
	if tweetProcessedBefore(tweet) {
		logAndPrint(fmt.Sprintf("Tweet proccessed before %s/status/%s", tweet.InReplyToScreenName, tweet.InReplyToStatusIDStr))
		return
	}
	// let's make sure it has "share this" in the string
	if !strings.Contains(strings.ToLower(tweet.Text), "@shareaspic") {
		return
	}
	if !strings.Contains(strings.ToLower(tweet.Text), "share this") {
		replyWithIDoNotUnderstand(tweet)
		return
	}

	makeTweetPicAndShare(tweet)
}

func tweetProcessedBefore(tweet twitter.Tweet) bool {
	result, _ := redisClient.SAdd(ProcessedTweets, tweet.ID).Result()
	var processed bool
	processed = result == 0
	if !processed {
		log.Printf("new tweet to process: %s\n", tweet.IDStr)
	}

	return processed
}

func replyWithIDoNotUnderstand(tweet twitter.Tweet) {
	log.Printf("replyWithIDoNotUnderstand: %s\n", tweet.IDStr)
	statusUpdate := &twitter.StatusUpdateParams{
		Status:             "",
		InReplyToStatusID:  tweet.ID,
		PossiblySensitive:  nil,
		Lat:                nil,
		Long:               nil,
		PlaceID:            "",
		DisplayCoordinates: nil,
		TrimUser:           nil,
		MediaIds:           nil,
		TweetMode:          "",
	}
	_, _, err := client.Statuses.Update(fmt.Sprintf("Hello @%s , Sorry but I do not understand your message!", tweet.User.ScreenName), statusUpdate)
	if err != nil {
		logAndPrint(fmt.Sprintf("faild to reply with do not understand message %s", err.Error()))
	}
}

func makeTweetPicAndShare(tweet twitter.Tweet) {
	logAndPrint(fmt.Sprintf("prepare replyWithScreenShotFor: %s\n", tweet.IDStr))

	logAndPrint("taking a screenshot")
	filename, err := TweetScreenShot(tweet.InReplyToScreenName, tweet.InReplyToStatusIDStr)
	if err != nil {
		logAndPrint(fmt.Sprintf("Faild to take a screenshot of the tweet, %s", err.Error()))
		return
	}
	logAndPrint("screenshot has been taken successfully")

	logAndPrint(fmt.Sprintf("replying to %s (%s) for reply to %s/status/%s", tweet.User.ScreenName, tweet.IDStr, tweet.InReplyToScreenName, tweet.InReplyToStatusIDStr))

	filename = fmt.Sprintf("%s%s", PIC_STORAGE_PATH, filename)

	replyMessage := fmt.Sprintf("Hello @%s , here u are", tweet.User.ScreenName)

	err2 := TweetSendReply(tweet.User.ScreenName, tweet.IDStr, replyMessage, filename)
	if err2 != nil {
		logAndPrint(fmt.Sprintf("Faild to reply with a screenshot: %s", err2.Error()))
	}

	logAndPrint(fmt.Sprintf("replied With screenshot for: %s\n", tweet.IDStr))
}
