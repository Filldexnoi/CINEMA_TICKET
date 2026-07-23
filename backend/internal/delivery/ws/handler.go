package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type TokenVerifier interface {
	Verify(token string) (userID string, err error)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Handler(hub *Hub, verifier TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		showtimeID := c.Param("id")
		if _, err := verifier.Verify(c.Query("token")); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := NewClient(conn)
		hub.Register(showtimeID, client)

		go client.WritePump()
		client.ReadPump(func() {
			hub.Unregister(showtimeID, client)
		})
	}
}
