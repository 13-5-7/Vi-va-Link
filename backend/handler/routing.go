package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	"log"

	"github.com/labstack/echo/v4"
)

type RoutingHandler struct {
	osrmBaseURL string
}

func NewRoutingHandler(osrmBaseURL string) *RoutingHandler {
	return &RoutingHandler{osrmBaseURL: osrmBaseURL}
}

func (h *RoutingHandler) GetRoute(c echo.Context) error {
	log.Println("----routing GetRoute called-----")

	originLngStr := c.QueryParam("origin_lng")
	originLatStr := c.QueryParam("origin_lat")
	destLngStr := c.QueryParam("dest_lng")
	destLatStr := c.QueryParam("dest_lat")

	// 文字列からfloat64に変換
	originLng, err := strconv.ParseFloat(originLngStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid origin_lng"})
	}
	originLat, err := strconv.ParseFloat(originLatStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid origin_lat"})
	}
	destLng, err := strconv.ParseFloat(destLngStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid dest_lng"})
	}
	destLat, err := strconv.ParseFloat(destLatStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid dest_lat"})
	}

	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		h.osrmBaseURL, originLng, originLat, destLng, destLat)
	
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		if err == nil {
			defer resp.Body.Close()
			body, readErr := io.ReadAll(resp.Body)
			if readErr == nil && resp.StatusCode == http.StatusOK {
				var result interface {}
				if json.Unmarshal(body, &result) == nil {
					return c.JSON(http.StatusOK, result)
				}
			}
		}

		fallback := map[string]interface{}{
			"type": "LineString",
			"coordinates": [][]float64{
				{originLng, originLat},
				{destLng, destLat},
			},
		}
		return c.JSON(http.StatusOK, fallback)
}
