package web

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	apidocs "harns-platform/apis"
)

const swaggerPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Harns API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div style="padding:12px 16px;font-family:sans-serif">
    <label for="spec">OpenAPI spec:</label>
    <select id="spec" onchange="location.href='/swagger?spec='+encodeURIComponent(this.value)">%s</select>
  </div>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: '/api/v1/openapi/%s',
      dom_id: '#swagger-ui',
      deepLinking: true,
      displayRequestDuration: true
    });
  </script>
</body>
</html>`

func specNames() ([]string, error) {
	entries, err := apidocs.FS.ReadDir("platform")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func installOpenAPI(router *gin.Engine, v1 *gin.RouterGroup) {
	v1.GET("/openapi", func(c *gin.Context) {
		names, err := specNames()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 10000, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"specs": names}})
	})

	v1.GET("/openapi/:file", func(c *gin.Context) {
		name := c.Param("file")
		names, err := specNames()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 10000, "message": err.Error()})
			return
		}
		valid := false
		for _, n := range names {
			if n == name {
				valid = true
				break
			}
		}
		if !valid {
			c.JSON(http.StatusNotFound, gin.H{"code": 10004, "message": "spec not found"})
			return
		}
		data, err := apidocs.FS.ReadFile("platform/" + name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 10000, "message": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", data)
	})

	router.GET("/swagger", func(c *gin.Context) {
		names, err := specNames()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		if len(names) == 0 {
			c.String(http.StatusNotFound, "no OpenAPI specs found")
			return
		}
		selected := c.Query("spec")
		found := false
		var options strings.Builder
		for _, n := range names {
			sel := ""
			if n == selected {
				sel = " selected"
				found = true
			}
			fmt.Fprintf(&options, `<option value="%s"%s>%s</option>`, n, sel, n)
		}
		if !found {
			selected = names[0]
			options.Reset()
			for i, n := range names {
				sel := ""
				if i == 0 {
					sel = " selected"
				}
				fmt.Fprintf(&options, `<option value="%s"%s>%s</option>`, n, sel, n)
			}
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8",
			[]byte(fmt.Sprintf(swaggerPage, options.String(), selected)))
	})
}
