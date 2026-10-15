package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"google.golang.org/genai"

	"switchboard/internal/utils"
)

type PriceAdvisorRequest struct {
	Description        string  `json:"description" validate:"required"`
	Category           string  `json:"category" validate:"required"`
	CurrentAskingPrice float64 `json:"current_asking_price" validate:"required,gt=0"`
}

type PriceAdvisorResponse struct {
	FairPriceRange      string `json:"fair_price_range"`
	BetterTitle         string `json:"better_title"`
	ImprovedDescription string `json:"improved_description"`
}

func (handler *Handler) PriceAdvisor(writer http.ResponseWriter, req *http.Request) {
	var reqBody PriceAdvisorRequest
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := handler.validate.Struct(reqBody); err != nil {
		utils.SendError(writer, http.StatusBadRequest, err.Error())
		return
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")

	if geminiKey == "" || geminiKey == "mock-key" {
		mockResp := PriceAdvisorResponse{
			FairPriceRange:      fmt.Sprintf("$%.2f - $%.2f", reqBody.CurrentAskingPrice*0.85, reqBody.CurrentAskingPrice*1.15),
			BetterTitle:         fmt.Sprintf("Premium %s - %s", reqBody.Category, reqBody.Description),
			ImprovedDescription: fmt.Sprintf("Excellent quality %s. Description: %s", reqBody.Category, reqBody.Description),
		}
		utils.SendJSON(writer, http.StatusOK, mockResp)
		return
	}

	ctx := req.Context()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: geminiKey,
	})
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "failed to create AI client: "+err.Error())
		return
	}

	prompt := fmt.Sprintf(
		"Analyze the following item: Category: %s, Current Price: $%.2f, Description: %s. "+
			"Respond ONLY with a JSON object containing the keys 'fair_price_range', 'better_title', and 'improved_description'. "+
			"Do not include any markdown wrappers or backticks.",
		reqBody.Category, reqBody.CurrentAskingPrice, reqBody.Description,
	)

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), nil)
	if err != nil {
		utils.SendError(writer, http.StatusInternalServerError, "AI service failed: "+err.Error())
		return
	}

	responseText := resp.Text()
	if responseText == "" {
		utils.SendError(writer, http.StatusInternalServerError, "AI returned empty response")
		return
	}

	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var advisorResp PriceAdvisorResponse
	if err := json.Unmarshal([]byte(responseText), &advisorResp); err != nil {
		advisorResp = PriceAdvisorResponse{
			FairPriceRange:      fmt.Sprintf("$%.2f - $%.2f", reqBody.CurrentAskingPrice*0.9, reqBody.CurrentAskingPrice*1.1),
			BetterTitle:         "Suggested Title: " + reqBody.Category,
			ImprovedDescription: responseText,
		}
	}

	utils.SendJSON(writer, http.StatusOK, advisorResp)
}
