package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	bottalk "github.com/bottalk/go-plugin"
)

type productsResponse struct {
	Products []struct {
		ProductID        string `json:"productId"`
		ProductReference string `json:"referenceName"`
	} `json:"inSkillProducts"`
}

type btRequest struct {
	Token        string `json:"token"`
	UserID       string `json:"user"`
	ProductID    string `json:"productId"`
	ProductToken string `json:"productToken"`
	Message      string `json:"message"`
	Reference    string `json:"reference"`

	Input struct {
		Context struct {
			System struct {
				ApiEndpoint string `json:"apiEndpoint"`
				ApiToken    string `json:"apiAccessToken"`
			} `json:"System"`
		} `json:"context"`
		Request struct {
			Locale string `json:"locale"`
		} `json:"request"`
	} `json:"input"`
}

func errorResponse(message string) string {
	return "{\"result\": \"fail\",\"message\":\"" + message + "\"}"
}

func main() {

	plugin := bottalk.NewPlugin()
	plugin.Name = "ISP Plugin"
	plugin.Description = "This plugin helps you to make In-Skill-Purchases in alexa"

	plugin.Actions = map[string]bottalk.Action{
		"getProducts": {
			Name:        "getProducts",
			Description: "This queries all products",
			Endpoint:    "/getproducts",
			Action: func(r *http.Request) string {
				log.Println("Invoke getproducts")

				var BTR btRequest
				decoder := json.NewDecoder(r.Body)

				err := decoder.Decode(&BTR)
				if err != nil {
					return errorResponse(err.Error())
				}

				log.Println(BTR.Input.Context.System.ApiToken)

				// /v1/users/~current/skills/~current/inSkillProducts

				req, err := http.NewRequest(http.MethodGet, BTR.Input.Context.System.ApiEndpoint+"/v1/users/~current/skills/~current/inSkillProducts", nil)

				req.Header.Set("Authorization", "Bearer "+BTR.Input.Context.System.ApiToken)
				req.Header.Set("Accept-Language", BTR.Input.Request.Locale)

				log.Println(req)

				client := &http.Client{}

				res, err := client.Do(req)

				if err != nil {
					return errorResponse(err.Error())
				}
				output, _ := ioutil.ReadAll(res.Body)
				log.Println(string(output))

				return "{\"result\": \"ok\",\"response\":" + string(output) + "}"
			},
		},
		"getProduct": {
			Name:        "getProduct",
			Description: "This queries exact product by reference",
			Endpoint:    "/getproduct",
			Params:      map[string]string{"reference": "reference name of the product"},
			Action: func(r *http.Request) string {

				var BTR btRequest
				decoder := json.NewDecoder(r.Body)

				err := decoder.Decode(&BTR)
				if err != nil {
					return errorResponse(err.Error())
				}

				log.Println(BTR.Input.Context.System.ApiToken)

				req, err := http.NewRequest(http.MethodGet, BTR.Input.Context.System.ApiEndpoint+"/v1/users/~current/skills/~current/inSkillProducts", nil)

				req.Header.Set("Authorization", "Bearer "+BTR.Input.Context.System.ApiToken)
				req.Header.Set("Accept-Language", BTR.Input.Request.Locale)

				log.Println(req)

				client := &http.Client{}

				res, err := client.Do(req)

				if err != nil {
					return errorResponse(err.Error())
				}

				decoder = json.NewDecoder(res.Body)

				alexaResp := productsResponse{}
				err = decoder.Decode(&alexaResp)
				if err != nil {
					return errorResponse(err.Error())
				}

				productId := ""
				for _, product := range alexaResp.Products {
					if product.ProductReference == BTR.Reference {
						productId = product.ProductID
					}
				}

				if productId == "" {
					log.Println("Can't find product by reference: " + BTR.Reference)
					return errorResponse("Product not found")
				}

				req, err = http.NewRequest(http.MethodGet, BTR.Input.Context.System.ApiEndpoint+"/v1/users/~current/skills/~current/inSkillProducts/"+productId, nil)

				req.Header.Set("Authorization", "Bearer "+BTR.Input.Context.System.ApiToken)
				req.Header.Set("Accept-Language", BTR.Input.Request.Locale)

				client = &http.Client{}

				res, err = client.Do(req)

				if err != nil {
					return errorResponse(err.Error())
				}
				output, _ := ioutil.ReadAll(res.Body)
				log.Println(string(output))

				return "{\"result\": \"ok\",\"response\":" + string(output) + "}"
			},
		},
		"upsell": {
			Name:        "upsell",
			Description: "Performs upsell for request",
			Endpoint:    "/upsell",
			Params:      map[string]string{"productId": "ID of a product", "productToken": "Token to continue session", "message": "Upsell message"},
			Action: func(r *http.Request) string {
				log.Println("Invoke upsell")

				var BTR btRequest
				decoder := json.NewDecoder(r.Body)

				err := decoder.Decode(&BTR)
				if err != nil {
					return errorResponse(err.Error())
				}

				upsellRequest := `{
  "version": "1.0",
  "response": {
    "shouldEndSession": true,
    "directives": [
      {
        "type": "Connections.SendRequest",
        "name": "Upsell",
        "token": "` + BTR.ProductToken + `",
        "payload": {
          "InSkillProduct": {
            "productId": "` + BTR.ProductID + `"
          },
          "upsellMessage": "` + BTR.Message + `"
        }
      }
    ]
  },
  "sessionAttributes": {}
}`

				return "{\"result\": \"ok\",\"output\":" + upsellRequest + "}"
			},
		},
		"cancel": {
			Name:        "cancel",
			Description: "Performs cancel request",
			Endpoint:    "/cancel",
			Params:      map[string]string{"productId": "ID of a product", "productToken": "Token to continue session"},
			Action: func(r *http.Request) string {

				var BTR btRequest
				decoder := json.NewDecoder(r.Body)

				err := decoder.Decode(&BTR)
				if err != nil {
					return errorResponse(err.Error())
				}

				upsellRequest := `{
  "version": "1.0",
  "response": {
    "shouldEndSession": true,
    "directives": [
      {
        "type": "Connections.SendRequest",
        "name": "Cancel",
        "token": "` + BTR.ProductToken + `",
        "payload": {
          "InSkillProduct": {
            "productId": "` + BTR.ProductID + `"
          }
        }
      }
    ]
  },
  "sessionAttributes": {}
}`

				return "{\"result\": \"ok\",\"output\":" + upsellRequest + "}"
			},
		},

		"buy": {
			Name:        "buy",
			Description: "Performs buy request",
			Endpoint:    "/buy",
			Params:      map[string]string{"productId": "ID of a product", "productToken": "Token to continue session"},
			Action: func(r *http.Request) string {

				var BTR btRequest
				decoder := json.NewDecoder(r.Body)

				err := decoder.Decode(&BTR)
				if err != nil {
					return errorResponse(err.Error())
				}

				upsellRequest := `{
  "version": "1.0",
  "response": {
    "shouldEndSession": true,
    "directives": [
      {
        "type": "Connections.SendRequest",
        "name": "Buy",
        "token": "` + BTR.ProductToken + `",
        "payload": {
          "InSkillProduct": {
            "productId": "` + BTR.ProductID + `"
          }
        }
      }
    ]
  },
  "sessionAttributes": {}
}`

				return "{\"result\": \"ok\",\"output\":" + upsellRequest + "}"
			},
		},
	}

	plugin.Run(":9065")
}
