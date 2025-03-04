package api_v1

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/bata94/RegattaApi/internal/crud"
	"github.com/bata94/RegattaApi/internal/handlers"
	"github.com/bata94/RegattaApi/internal/handlers/api"
	"github.com/bata94/RegattaApi/internal/sqlc"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type WSZnMsg struct {
	Status *string         `json:"status"`
	Method string          `json:"method"`
	Data   *crud.Zeitnahme `json:"data"`
}

func WsZeitnahmeZiel(c *websocket.Conn) {
	// When the function returns, unregister the client and close the connection
	defer func() {
		handlers.Unregister <- c
		c.Close()
	}()

	// Register the client
	handlers.Register <- c

	q, err := crud.GetOpenZeitnahmeZiel()
	if err != nil {
		errStr := fmt.Sprint("Error getting open ZnZiel... ", err)
		log.Error(errStr)
		c.WriteMessage(1, []byte(errStr))
		return
	}

	qJson, err := json.Marshal(fiber.Map{
		"list": q,
	})
	if err != nil {
		errStr := fmt.Sprint("Error getting open ZnZiel... ", err)
		log.Error(errStr)
		c.WriteMessage(1, []byte(errStr))
		return
	}
	c.WriteMessage(1, qJson)

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("read error:", err)
			}

			return // Calls the deferred function, i.e. closes the connection on error
		}

		if messageType == websocket.TextMessage {
			retMsg := ""
			var msg WSZnMsg
			json.Unmarshal(message, &msg)

			if msg.Method == "post" {
				if msg.Data == nil || msg.Data.TimeClient == nil || msg.Data.MeasuredLatency == nil {
					retMsg = "Bad Request: TimeClient or MeasuredLatency is nil or unparsable"
					goto ReturnMessage
				}

				q, err := crud.CreateZeitnahmeZiel(nil, nil, *msg.Data.TimeClient, *msg.Data.MeasuredLatency)
				if err != nil {
					retMsg = "Error:" + err.Error()
					goto ReturnMessage
				}
				qJson, err := json.Marshal(fiber.Map{
					"new": q,
				})
				if err != nil {
					retMsg = "Error:" + err.Error()
					goto ReturnMessage
				} else {
					retMsg = string(qJson)
				}
			} else if msg.Method == "put" {
				if msg.Data == nil {
					retMsg = "Bad Request: TimeClient or MeasuredLatency is nil or unparsable"
					goto ReturnMessage
				}

				zeitnahme, err := crud.GetZeitnahmeZiel(int(msg.Data.ID))
				if err != nil {
					retMsg = "Error: " + err.Error()
					goto ReturnMessage
				}

				log.Debug(zeitnahme, msg.Data)
				q, err := crud.UpdateZeitnahmeZiel(zeitnahme, msg.Data.RennenNummer, msg.Data.StartNummer)
				if err != nil {
					retMsg = "Error: " + err.Error()
					goto ReturnMessage
				}

				qJson, err := json.Marshal(fiber.Map{
					"update": q,
				})

				c.WriteMessage(1, qJson)
			} else if msg.Method == "delete" {
				if msg.Data == nil {
					retMsg = "Bad Request: ID is nil or unparsable"
					goto ReturnMessage
				}

				z, err := crud.GetZeitnahmeZiel(int(msg.Data.ID))
				if err != nil {
					retMsg = "Error:" + err.Error()
					goto ReturnMessage
				}
				q, err := crud.DeleteZeitnahmeZiel(z)
				if err != nil {
					retMsg = "Error:" + err.Error()
					goto ReturnMessage
				}
				qJson, err := json.Marshal(fiber.Map{
					"delete": q,
				})

				c.WriteMessage(1, qJson)
			} else if msg.Method == "get" {
				q, err := crud.GetOpenZeitnahmeZiel()
				if err != nil {
					errStr := fmt.Sprint("Error getting open ZnZiel... ", err)
					log.Error(errStr)
					c.WriteMessage(1, []byte(errStr))
					return
				}

				qJson, err := json.Marshal(fiber.Map{
					"list": q,
				})
				if err != nil {
					errStr := fmt.Sprint("Error getting open ZnZiel... ", err)
					log.Error(errStr)
					c.WriteMessage(1, []byte(errStr))
					return
				}
				c.WriteMessage(1, qJson)
			} else if msg.Method == "ping" {
				c.WriteMessage(1, []byte("pong"))
			}

		ReturnMessage:
			if retMsg != "" {
				handlers.Broadcast <- retMsg
			}
		} else {
			log.Error("websocket message received of type", messageType)
		}
	}
}

type PostStartParams struct {
	RennenNummer    *string   `json:"renn_nummer"`
	StartNummern    []string  `json:"start_nummern"`
	TimeClient      time.Time `json:"time_client"`
	MeasuredLatency *int      `json:"measured_latency"`
}

func PostZeitnahmeStart(c *fiber.Ctx) error {
	p := new(PostStartParams)
	err := c.BodyParser(p)
	if err != nil {
		return err
	}

	q, err := crud.CreateZeitnahmeStart(p.RennenNummer, p.StartNummern, p.TimeClient, *p.MeasuredLatency)
	if err != nil {
		return err
	}

	return api.JSON(c, q)
}

func GetOpenStarts(c *fiber.Ctx) error {
	q, err := crud.GetOpenZeitnahmeStart()
	if err != nil {
		return err
	}

	return api.JSON(c, q)
}

func GenerateEndZeit(c *fiber.Ctx) error {
	starts, err := crud.GetOpenZeitnahmeStart()
	if err != nil {
		log.Error("GetOpenZeitnahmeStart")
		return err
	}

	ziels, err := crud.GetOpenZeitnahmeZiel()
	if err != nil {
		log.Error("GetOpenZeitnahmeZiel")
		return err
	}

	if len(ziels) == 0 {
		log.Error("0 Ziels")
		return &api.BAD_REQUEST
	}
	if len(starts) == 0 {
		log.Error("0 Starts")
		return &api.BAD_REQUEST
	}

	for _, z := range ziels {
		if z.StartNummer == nil || *z.StartNummer == "" {
			continue
		}
		for _, s := range starts {
			if *z.StartNummer == *s.StartNummer {
				startNummerInt, err := strconv.Atoi(*s.StartNummer)
				if err != nil {
					log.Error("Error StartNummerStr to int")
					return err
				}
				// TODO: Make Tag dynamic
				meld, err := crud.GetMeldungByStartNrUndTag(startNummerInt, sqlc.TagSa)
				if err != nil {
					log.Error("GetMeldungByStartNrUndTag")
					return err
				}
				if meld.Uuid == uuid.Nil {
					log.Error("GetMeldungByStartNrUndTag meld.Uuid == nil")
					return &api.BAD_REQUEST
				}

				err = crud.CreateZeitnahmeErgebnis(s, z, meld)
				if err != nil {
					log.Error("CreateZeitnahmeErgebnis")
					return err
				}
			}
		}
	}

	return api.JSON(c, "success")
}
