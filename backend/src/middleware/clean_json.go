package middleware

import (
	"bytes"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

// CleanJSON remove caracteres invisíveis e BOM do body JSON
func CleanJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "multipart/form-data") {
			c.Next()
			return
		}
		if c.GetHeader("Stripe-Signature") != "" {
			c.Next()
			return
		}
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" || c.Request.Method == "GET" {
			body, _ := io.ReadAll(c.Request.Body)
			body = cleanJSONBody(body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		c.Next()
	}
}

func cleanJSONBody(body []byte) []byte {
	// Remove caracteres invisíveis problemáticos
	removals := [][]byte{
		[]byte{0xC2, 0xA0},       // NO-BREAK SPACE
		[]byte{0xE2, 0x80, 0x8B}, // Zero Width Space
		[]byte{0xE2, 0x80, 0x8C}, // Zero Width Non-Joiner
		[]byte{0xE2, 0x80, 0x8D}, // Zero Width Joiner
		[]byte{0xE2, 0x80, 0x8E}, // Left-to-Right Mark
		[]byte{0xE2, 0x80, 0x8F}, // Right-to-Left Mark
		[]byte{0xEF, 0xBB, 0xBF}, // BOM UTF-8
	}
	for _, pattern := range removals {
		body = bytes.ReplaceAll(body, pattern, []byte{})
	}
	// Remove ASCII control chars: \r, \t, \v, \f (se necessário)
	body = bytes.ReplaceAll(body, []byte{0x0D}, []byte{}) // \r
	body = bytes.ReplaceAll(body, []byte{0x09}, []byte{}) // \t
	body = bytes.ReplaceAll(body, []byte{0x0B}, []byte{}) // \v
	body = bytes.ReplaceAll(body, []byte{0x0C}, []byte{}) // \f
	return body
}
