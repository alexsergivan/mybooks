package renderer

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinifyHTML(t *testing.T) {
	b, _ := ioutil.ReadFile("mocks/example-template.html")
	expected := `{{ define "city-connection-stops" }}{{ $cityFrom := (index .Data.Page.TranslationOf.Places 0) }}{{ $cityTo := (index .Data.Page.TranslationOf.Places 1) }}<section id="stops-location" class="flix-page-container city-connection-stops-container"><nav class="flix-nav-horizontal" data-tabs><ul class="flix-nav-horizontal__items city-connection-stops-nav"><li class="flix-nav-horizontal__item"><a class="flix-nav-horizontal__link" href="#first-city-content" data-tab-active-class="flix-nav-horizontal__link--active">{{ if eq $cityFrom.Content.Name "" }}{{ $cityFrom.Name }}{{ else }}{{$cityFrom.Content.Name}}{{ end }}</a></li><li class="flix-nav-horizontal__item"><a class="flix-nav-horizontal__link" href="#second-city-content" data-tab-active-class="flix-nav-horizontal__link--active">{{ if eq $cityTo.Content.Name "" }}{{ $cityTo.Name }}{{ else }}{{$cityTo.Content.Name}}{{ end }}</a></li></ul></nav><div class="city-connection-tab-content" data-tabs-content><div id="first-city-content">{{ template "stops-location" args "city" $cityFrom "translations" .AllTranslations }}</div><div style="display: none" id="second-city-content">{{ template "stops-location" args "city" $cityTo "translations" .AllTranslations }}</div></div></section><script>const mapElement = document.querySelector("#stops-location");lazyInit(mapElement, function() {const leafletStyle = document.createElement("link");leafletStyle.rel = "stylesheet";leafletStyle.href = "https://unpkg.com/leaflet@1.7.1/dist/leaflet.css";leafletStyle.integrity = "sha512-xodZBNTC5n17Xt2atTPuE1HxjVMSvLVW9ocqUKLsCC5CXdbqCmblAshOMAS6/keqq/sMZMZ19scR4PsZChSR7A==";leafletStyle.crossOrigin = "";document.head.appendChild(leafletStyle);const leafletScript = document.createElement('script');leafletScript.src = "https://unpkg.com/leaflet@1.7.1/dist/leaflet.js";leafletScript.integrity = "sha512-XQoYMqMTK8LvdxXYG3nZ448hOEQiglfqkJs1NOQV44cWnUrBc8PkAOcXy20w0vlaXaVUearIOBhiXZ5V3ynxwA==";leafletScript.crossOrigin = "";leafletScript.onload = function leafletOnLoad() {try {handleCityConnectionStopsLocation({{ $cityFrom.Has }}, {{ $cityTo.Has }}, {{$cityFrom.Slug}}, {{$cityTo.Slug}});} catch (e) {console.log("Error during stops map initialisation");}};document.head.appendChild(leafletScript);});</script>{{ end }}`
	minifiedHtml, _ := MinifyHTML(b)
	assert.Equal(t, expected, minifiedHtml)
}
