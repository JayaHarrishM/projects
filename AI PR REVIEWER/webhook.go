package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func verifySignature(body []byte, signature string) bool {
	secret := os.Getenv("WEBHOOK_SECRET")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Got a request !.\nMethod: ", r.Method)
	if r.Method != http.MethodPost {
		http.Error(w, "Thanks for using the site..", http.StatusOK)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("cannot read body: ", http.StatusBadRequest)
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	log.Println("Got the webhook request body.")
	signature := r.Header.Get("X-Hub-Signature-256")
	if !verifySignature(body, signature) {
		log.Println("invalid signature: ", http.StatusUnauthorized)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	log.Println("Signature test passed.")
	var event PullRequestEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Println("invalid json: ", http.StatusBadRequest)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	log.Println("Pullrequest event: ", event)
	if event.Action != "opened" && event.Action != "reopened" {
		log.Println("Confirming this is not an open or reopened pr..")
		w.WriteHeader(http.StatusOK)
		return
	}

	go processPR(event.Repository.FullName, event.PullRequest.Number)

	w.WriteHeader(http.StatusOK)
}

func processPR(repo string, prNumber int) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	diff, err := GetPRDiff(ctx, repo, prNumber)
	if err != nil {
		log.Println("Failed to get pr diff: ", err)
		return
	}
	log.Println("\n\n\nThe PR diff is: ", diff)

	review, err := ReviewCode(ctx, diff)
	if err != nil {
		log.Println("Failed to review the code", err)
		return
	}

	err = PostComment(ctx, repo, prNumber, review)
	if err != nil {
		log.Println("Failed to add a comment: ", err)
	}
}
