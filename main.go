package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type WindSpeedBody struct {
	Lat  string `json:"lat"`
	Long string `json:"long"`
}

type WindSpeedResponse struct {
	Result string `json:"result"`
}

type WeatherResponse struct {
	Latitude         float64        `json:"latitude"`
	Longitude        float64        `json:"longitude"`
	GenerationTimeMs float64        `json:"generationtime_ms"`
	UtcOffsetSeconds int            `json:"utc_offset_seconds"`
	Timezone         string         `json:"timezone"`
	TimezoneAbbr     string         `json:"timezone_abbreviation"`
	Elevation        float64        `json:"elevation"`
	CurrentUnits     CurrentUnits   `json:"current_units"`
	Current          CurrentWeather `json:"current"`
}

type CurrentUnits struct {
	Time         string `json:"time"`
	Interval     string `json:"interval"`
	WindSpeed10m string `json:"wind_speed_10m"`
}

type CurrentWeather struct {
	Time         string  `json:"time"`
	Interval     int     `json:"interval"`
	WindSpeed10m float64 `json:"wind_speed_10m"`
}

func main() {
	http.HandleFunc("/", windSpeedHandler)
	http.ListenAndServe(":8080", nil)
}

func windSpeedHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var body WindSpeedBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	lat := body.Lat
	long := body.Long

	thirdPartyResponse, err := http.Get(getWeatherApiString(lat, long))
	if err != nil {
		http.Error(w, "Third party api did not respond", http.StatusBadRequest)
		return
	}

	var thirdPartyParsed WeatherResponse
	if err := json.NewDecoder(thirdPartyResponse.Body).Decode(&thirdPartyParsed); err != nil {
		http.Error(w, "There was an error while parsing the weather api response", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	myResponse := WindSpeedResponse{Result: "The wind speed in your location is " +
		fmt.Sprint(thirdPartyParsed.Current.WindSpeed10m) + fmt.Sprint(thirdPartyParsed.CurrentUnits.WindSpeed10m)}

	marshaledResponse, err := json.Marshal(myResponse)
	if err != nil {
		http.Error(w, "There was an error whle marshaling the response", http.StatusBadRequest)
	}

	w.Write(marshaledResponse)
}

func getWeatherApiString(lat string, long string) string {
	windSpeedApiBaseUrl := "https://api.open-meteo.com/v1/forecast"
	parsedApiUrl, _ := url.Parse(windSpeedApiBaseUrl)
	q := parsedApiUrl.Query()
	q.Set("latitude", lat)
	q.Set("longitude", long)
	q.Set("current", "wind_speed_10m")
	parsedApiUrl.RawQuery = q.Encode()
	return parsedApiUrl.String()
}
