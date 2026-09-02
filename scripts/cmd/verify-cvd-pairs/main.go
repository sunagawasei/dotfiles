package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/sunagawasei/dotfiles/scripts/internal/cvd"
	"github.com/sunagawasei/dotfiles/scripts/internal/palette"
	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

const cvdThreshold = 20.0

const legacyFixtureSHA256 = "f52668fac28ee7b873ebc19d272ef07f4377a301c2bfbf1b0cfc0f85601a4c41"

type checkKind string

const (
	KindHistoricalFloor   checkKind = "historical_floor"
	KindAchievedFreeze    checkKind = "achieved_freeze"
	KindMinimumSeparation checkKind = "minimum_separation"
)

type check struct {
	First      string
	Second     string
	Vision     string
	BaselineDE float64
	Kind       checkKind
}

type waiver struct {
	First      string
	Second     string
	Vision     string
	ApprovedDE float64
}

var ansiChecks = []check{
	{First: "ansi.black", Second: "ansi.red", Vision: "protanopia", BaselineDE: 48.0143},
	{First: "ansi.black", Second: "ansi.red", Vision: "deuteranopia", BaselineDE: 54.8045},
	{First: "ansi.black", Second: "ansi.red", Vision: "tritanopia", BaselineDE: 66.6722},
	{First: "ansi.black", Second: "ansi.green", Vision: "protanopia", BaselineDE: 54.9232},
	{First: "ansi.black", Second: "ansi.green", Vision: "deuteranopia", BaselineDE: 50.3148},
	{First: "ansi.black", Second: "ansi.green", Vision: "tritanopia", BaselineDE: 60.7175},
	{First: "ansi.black", Second: "ansi.yellow", Vision: "protanopia", BaselineDE: 69.1789},
	{First: "ansi.black", Second: "ansi.yellow", Vision: "deuteranopia", BaselineDE: 72.3334},
	{First: "ansi.black", Second: "ansi.yellow", Vision: "tritanopia", BaselineDE: 65.7185},
	{First: "ansi.black", Second: "ansi.blue", Vision: "protanopia", BaselineDE: 57.8547},
	{First: "ansi.black", Second: "ansi.blue", Vision: "deuteranopia", BaselineDE: 56.7773},
	{First: "ansi.black", Second: "ansi.blue", Vision: "tritanopia", BaselineDE: 56.4841},
	{First: "ansi.black", Second: "ansi.magenta", Vision: "protanopia", BaselineDE: 51.0495},
	{First: "ansi.black", Second: "ansi.magenta", Vision: "deuteranopia", BaselineDE: 54.4574},
	{First: "ansi.black", Second: "ansi.magenta", Vision: "tritanopia", BaselineDE: 75.2545},
	{First: "ansi.black", Second: "ansi.cyan", Vision: "protanopia", BaselineDE: 71.2334},
	{First: "ansi.black", Second: "ansi.cyan", Vision: "deuteranopia", BaselineDE: 68.4939},
	{First: "ansi.black", Second: "ansi.cyan", Vision: "tritanopia", BaselineDE: 76.9433},
	{First: "ansi.black", Second: "ansi.white", Vision: "protanopia", BaselineDE: 81.3067},
	{First: "ansi.black", Second: "ansi.white", Vision: "deuteranopia", BaselineDE: 80.5502},
	{First: "ansi.black", Second: "ansi.white", Vision: "tritanopia", BaselineDE: 80.8961},
	{First: "ansi.black", Second: "ansi.bright_black", Vision: "protanopia", BaselineDE: 29.2795},
	{First: "ansi.black", Second: "ansi.bright_black", Vision: "deuteranopia", BaselineDE: 30.7941},
	{First: "ansi.black", Second: "ansi.bright_black", Vision: "tritanopia", BaselineDE: 30.1900},
	{First: "ansi.black", Second: "ansi.bright_red", Vision: "protanopia", BaselineDE: 52.2597},
	{First: "ansi.black", Second: "ansi.bright_red", Vision: "deuteranopia", BaselineDE: 59.3896},
	{First: "ansi.black", Second: "ansi.bright_red", Vision: "tritanopia", BaselineDE: 70.5790},
	{First: "ansi.black", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 66.1567},
	{First: "ansi.black", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 61.4585},
	{First: "ansi.black", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 72.2830},
	{First: "ansi.black", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 77.3407},
	{First: "ansi.black", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 81.9475},
	{First: "ansi.black", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 74.5330},
	{First: "ansi.black", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 63.4159},
	{First: "ansi.black", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 63.1560},
	{First: "ansi.black", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 63.0193},
	{First: "ansi.black", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 72.6885},
	{First: "ansi.black", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 74.5839},
	{First: "ansi.black", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 76.7049},
	{First: "ansi.black", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 76.0137},
	{First: "ansi.black", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 73.8153},
	{First: "ansi.black", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 77.7238},
	{First: "ansi.black", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 89.0450},
	{First: "ansi.black", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 89.9267},
	{First: "ansi.black", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 88.7623},
	{First: "ansi.red", Second: "ansi.green", Vision: "protanopia", BaselineDE: 10.3688},
	{First: "ansi.red", Second: "ansi.green", Vision: "deuteranopia", BaselineDE: 8.5986},
	{First: "ansi.red", Second: "ansi.green", Vision: "tritanopia", BaselineDE: 72.7129},
	{First: "ansi.red", Second: "ansi.yellow", Vision: "protanopia", BaselineDE: 33.3101},
	{First: "ansi.red", Second: "ansi.yellow", Vision: "deuteranopia", BaselineDE: 24.9315},
	{First: "ansi.red", Second: "ansi.yellow", Vision: "tritanopia", BaselineDE: 24.9470},
	{First: "ansi.red", Second: "ansi.blue", Vision: "protanopia", BaselineDE: 27.6255},
	{First: "ansi.red", Second: "ansi.blue", Vision: "deuteranopia", BaselineDE: 40.3862},
	{First: "ansi.red", Second: "ansi.blue", Vision: "tritanopia", BaselineDE: 67.2540},
	{First: "ansi.red", Second: "ansi.magenta", Vision: "protanopia", BaselineDE: 22.6933},
	{First: "ansi.red", Second: "ansi.magenta", Vision: "deuteranopia", BaselineDE: 17.9826},
	{First: "ansi.red", Second: "ansi.magenta", Vision: "tritanopia", BaselineDE: 13.0513},
	{First: "ansi.red", Second: "ansi.cyan", Vision: "protanopia", BaselineDE: 27.7840},
	{First: "ansi.red", Second: "ansi.cyan", Vision: "deuteranopia", BaselineDE: 35.6286},
	{First: "ansi.red", Second: "ansi.cyan", Vision: "tritanopia", BaselineDE: 80.1276},
	{First: "ansi.red", Second: "ansi.white", Vision: "protanopia", BaselineDE: 33.3122},
	{First: "ansi.red", Second: "ansi.white", Vision: "deuteranopia", BaselineDE: 28.0302},
	{First: "ansi.red", Second: "ansi.white", Vision: "tritanopia", BaselineDE: 57.0700},
	{First: "ansi.red", Second: "ansi.bright_black", Vision: "protanopia", BaselineDE: 20.5220},
	{First: "ansi.red", Second: "ansi.bright_black", Vision: "deuteranopia", BaselineDE: 24.3032},
	{First: "ansi.red", Second: "ansi.bright_black", Vision: "tritanopia", BaselineDE: 40.4141},
	{First: "ansi.red", Second: "ansi.bright_red", Vision: "protanopia", BaselineDE: 4.5289},
	{First: "ansi.red", Second: "ansi.bright_red", Vision: "deuteranopia", BaselineDE: 4.6751},
	{First: "ansi.red", Second: "ansi.bright_red", Vision: "tritanopia", BaselineDE: 4.3350},
	{First: "ansi.red", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 19.7152},
	{First: "ansi.red", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 10.3333},
	{First: "ansi.red", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 78.5982},
	{First: "ansi.red", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 41.7433},
	{First: "ansi.red", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 34.8083},
	{First: "ansi.red", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 22.5638},
	{First: "ansi.red", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 17.4036},
	{First: "ansi.red", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 20.0853},
	{First: "ansi.red", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 45.5686},
	{First: "ansi.red", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 27.6121},
	{First: "ansi.red", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 27.4920},
	{First: "ansi.red", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 30.2899},
	{First: "ansi.red", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 28.6904},
	{First: "ansi.red", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 28.0266},
	{First: "ansi.red", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 68.9884},
	{First: "ansi.red", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 41.4537},
	{First: "ansi.red", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 35.5004},
	{First: "ansi.red", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 52.1937},
	{First: "ansi.green", Second: "ansi.yellow", Vision: "protanopia", BaselineDE: 23.2668},
	{First: "ansi.green", Second: "ansi.yellow", Vision: "deuteranopia", BaselineDE: 33.5275},
	{First: "ansi.green", Second: "ansi.yellow", Vision: "tritanopia", BaselineDE: 51.5649},
	{First: "ansi.green", Second: "ansi.blue", Vision: "protanopia", BaselineDE: 35.8735},
	{First: "ansi.green", Second: "ansi.blue", Vision: "deuteranopia", BaselineDE: 32.2580},
	{First: "ansi.green", Second: "ansi.blue", Vision: "tritanopia", BaselineDE: 10.6465},
	{First: "ansi.green", Second: "ansi.magenta", Vision: "protanopia", BaselineDE: 31.9801},
	{First: "ansi.green", Second: "ansi.magenta", Vision: "deuteranopia", BaselineDE: 10.8017},
	{First: "ansi.green", Second: "ansi.magenta", Vision: "tritanopia", BaselineDE: 85.2026},
	{First: "ansi.green", Second: "ansi.cyan", Vision: "protanopia", BaselineDE: 30.3608},
	{First: "ansi.green", Second: "ansi.cyan", Vision: "deuteranopia", BaselineDE: 29.7481},
	{First: "ansi.green", Second: "ansi.cyan", Vision: "tritanopia", BaselineDE: 17.3231},
	{First: "ansi.green", Second: "ansi.white", Vision: "protanopia", BaselineDE: 27.8269},
	{First: "ansi.green", Second: "ansi.white", Vision: "deuteranopia", BaselineDE: 30.2857},
	{First: "ansi.green", Second: "ansi.white", Vision: "tritanopia", BaselineDE: 36.8416},
	{First: "ansi.green", Second: "ansi.bright_black", Vision: "protanopia", BaselineDE: 25.8117},
	{First: "ansi.green", Second: "ansi.bright_black", Vision: "deuteranopia", BaselineDE: 21.5289},
	{First: "ansi.green", Second: "ansi.bright_black", Vision: "tritanopia", BaselineDE: 45.8464},
	{First: "ansi.green", Second: "ansi.bright_red", Vision: "protanopia", BaselineDE: 6.9166},
	{First: "ansi.green", Second: "ansi.bright_red", Vision: "deuteranopia", BaselineDE: 12.4373},
	{First: "ansi.green", Second: "ansi.bright_red", Vision: "tritanopia", BaselineDE: 73.7499},
	{First: "ansi.green", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 11.2640},
	{First: "ansi.green", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 11.1456},
	{First: "ansi.green", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 11.6090},
	{First: "ansi.green", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 31.5275},
	{First: "ansi.green", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 43.3711},
	{First: "ansi.green", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 61.1323},
	{First: "ansi.green", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 19.4548},
	{First: "ansi.green", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 16.0836},
	{First: "ansi.green", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 30.9986},
	{First: "ansi.green", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 28.8069},
	{First: "ansi.green", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 26.1707},
	{First: "ansi.green", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 56.6961},
	{First: "ansi.green", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 26.7092},
	{First: "ansi.green", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 26.0705},
	{First: "ansi.green", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 23.5298},
	{First: "ansi.green", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 34.3824},
	{First: "ansi.green", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 39.8623},
	{First: "ansi.green", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 50.9708},
	{First: "ansi.yellow", Second: "ansi.blue", Vision: "protanopia", BaselineDE: 58.6581},
	{First: "ansi.yellow", Second: "ansi.blue", Vision: "deuteranopia", BaselineDE: 64.5730},
	{First: "ansi.yellow", Second: "ansi.blue", Vision: "tritanopia", BaselineDE: 47.5847},
	{First: "ansi.yellow", Second: "ansi.magenta", Vision: "protanopia", BaselineDE: 55.1861},
	{First: "ansi.yellow", Second: "ansi.magenta", Vision: "deuteranopia", BaselineDE: 41.8781},
	{First: "ansi.yellow", Second: "ansi.magenta", Vision: "tritanopia", BaselineDE: 36.9886},
	{First: "ansi.yellow", Second: "ansi.cyan", Vision: "protanopia", BaselineDE: 48.9469},
	{First: "ansi.yellow", Second: "ansi.cyan", Vision: "deuteranopia", BaselineDE: 56.6132},
	{First: "ansi.yellow", Second: "ansi.cyan", Vision: "tritanopia", BaselineDE: 57.0268},
	{First: "ansi.yellow", Second: "ansi.white", Vision: "protanopia", BaselineDE: 34.3463},
	{First: "ansi.yellow", Second: "ansi.white", Vision: "deuteranopia", BaselineDE: 34.9931},
	{First: "ansi.yellow", Second: "ansi.white", Vision: "tritanopia", BaselineDE: 33.0116},
	{First: "ansi.yellow", Second: "ansi.bright_black", Vision: "protanopia", BaselineDE: 41.1775},
	{First: "ansi.yellow", Second: "ansi.bright_black", Vision: "deuteranopia", BaselineDE: 42.5110},
	{First: "ansi.yellow", Second: "ansi.bright_black", Vision: "tritanopia", BaselineDE: 35.5760},
	{First: "ansi.yellow", Second: "ansi.bright_red", Vision: "protanopia", BaselineDE: 30.1760},
	{First: "ansi.yellow", Second: "ansi.bright_red", Vision: "deuteranopia", BaselineDE: 21.7003},
	{First: "ansi.yellow", Second: "ansi.bright_red", Vision: "tritanopia", BaselineDE: 24.3063},
	{First: "ansi.yellow", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 21.3093},
	{First: "ansi.yellow", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 29.7745},
	{First: "ansi.yellow", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 55.6850},
	{First: "ansi.yellow", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 8.7314},
	{First: "ansi.yellow", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 10.1974},
	{First: "ansi.yellow", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 10.4653},
	{First: "ansi.yellow", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 39.3807},
	{First: "ansi.yellow", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 40.6170},
	{First: "ansi.yellow", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 23.1132},
	{First: "ansi.yellow", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 46.0993},
	{First: "ansi.yellow", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 42.2154},
	{First: "ansi.yellow", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 14.2212},
	{First: "ansi.yellow", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 40.2217},
	{First: "ansi.yellow", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 43.7384},
	{First: "ansi.yellow", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 45.0878},
	{First: "ansi.yellow", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 34.3983},
	{First: "ansi.yellow", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 33.5413},
	{First: "ansi.yellow", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 30.3539},
	{First: "ansi.blue", Second: "ansi.magenta", Vision: "protanopia", BaselineDE: 7.0893},
	{First: "ansi.blue", Second: "ansi.magenta", Vision: "deuteranopia", BaselineDE: 22.7296},
	{First: "ansi.blue", Second: "ansi.magenta", Vision: "tritanopia", BaselineDE: 79.2286},
	{First: "ansi.blue", Second: "ansi.cyan", Vision: "protanopia", BaselineDE: 20.0514},
	{First: "ansi.blue", Second: "ansi.cyan", Vision: "deuteranopia", BaselineDE: 16.9406},
	{First: "ansi.blue", Second: "ansi.cyan", Vision: "tritanopia", BaselineDE: 21.0987},
	{First: "ansi.blue", Second: "ansi.white", Vision: "protanopia", BaselineDE: 41.9620},
	{First: "ansi.blue", Second: "ansi.white", Vision: "deuteranopia", BaselineDE: 45.0086},
	{First: "ansi.blue", Second: "ansi.white", Vision: "tritanopia", BaselineDE: 35.4179},
	{First: "ansi.blue", Second: "ansi.bright_black", Vision: "protanopia", BaselineDE: 40.6742},
	{First: "ansi.blue", Second: "ansi.bright_black", Vision: "deuteranopia", BaselineDE: 42.6110},
	{First: "ansi.blue", Second: "ansi.bright_black", Vision: "tritanopia", BaselineDE: 40.4704},
	{First: "ansi.blue", Second: "ansi.bright_red", Vision: "protanopia", BaselineDE: 29.1463},
	{First: "ansi.blue", Second: "ansi.bright_red", Vision: "deuteranopia", BaselineDE: 42.9319},
	{First: "ansi.blue", Second: "ansi.bright_red", Vision: "tritanopia", BaselineDE: 68.4546},
	{First: "ansi.blue", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 39.3202},
	{First: "ansi.blue", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 35.4301},
	{First: "ansi.blue", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 18.5892},
	{First: "ansi.blue", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 66.2864},
	{First: "ansi.blue", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 73.7302},
	{First: "ansi.blue", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 57.3742},
	{First: "ansi.blue", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 21.6873},
	{First: "ansi.blue", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 26.2400},
	{First: "ansi.blue", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 26.3012},
	{First: "ansi.blue", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 23.6323},
	{First: "ansi.blue", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 33.5369},
	{First: "ansi.blue", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 51.8793},
	{First: "ansi.blue", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 31.5290},
	{First: "ansi.blue", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 31.6431},
	{First: "ansi.blue", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 24.2261},
	{First: "ansi.blue", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 51.8269},
	{First: "ansi.blue", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 57.2954},
	{First: "ansi.blue", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 49.0168},
	{First: "ansi.magenta", Second: "ansi.cyan", Vision: "protanopia", BaselineDE: 23.3992},
	{First: "ansi.magenta", Second: "ansi.cyan", Vision: "deuteranopia", BaselineDE: 19.1212},
	{First: "ansi.magenta", Second: "ansi.cyan", Vision: "tritanopia", BaselineDE: 92.0806},
	{First: "ansi.magenta", Second: "ansi.white", Vision: "protanopia", BaselineDE: 42.8938},
	{First: "ansi.magenta", Second: "ansi.white", Vision: "deuteranopia", BaselineDE: 28.2485},
	{First: "ansi.magenta", Second: "ansi.white", Vision: "tritanopia", BaselineDE: 67.9159},
	{First: "ansi.magenta", Second: "ansi.bright_black", Vision: "protanopia", BaselineDE: 33.8057},
	{First: "ansi.magenta", Second: "ansi.bright_black", Vision: "deuteranopia", BaselineDE: 29.2275},
	{First: "ansi.magenta", Second: "ansi.bright_black", Vision: "tritanopia", BaselineDE: 51.3268},
	{First: "ansi.magenta", Second: "ansi.bright_red", Vision: "protanopia", BaselineDE: 25.0715},
	{First: "ansi.magenta", Second: "ansi.bright_red", Vision: "deuteranopia", BaselineDE: 20.2217},
	{First: "ansi.magenta", Second: "ansi.bright_red", Vision: "tritanopia", BaselineDE: 12.9048},
	{First: "ansi.magenta", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 37.2771},
	{First: "ansi.magenta", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 13.2268},
	{First: "ansi.magenta", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 90.9376},
	{First: "ansi.magenta", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 63.2126},
	{First: "ansi.magenta", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 51.1232},
	{First: "ansi.magenta", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 32.5440},
	{First: "ansi.magenta", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 21.5421},
	{First: "ansi.magenta", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 8.7685},
	{First: "ansi.magenta", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 57.3910},
	{First: "ansi.magenta", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 26.3717},
	{First: "ansi.magenta", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 20.2221},
	{First: "ansi.magenta", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 39.2489},
	{First: "ansi.magenta", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 33.2575},
	{First: "ansi.magenta", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 19.3609},
	{First: "ansi.magenta", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 80.6088},
	{First: "ansi.magenta", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 52.6975},
	{First: "ansi.magenta", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 39.7471},
	{First: "ansi.magenta", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 61.5099},
	{First: "ansi.cyan", Second: "ansi.white", Vision: "protanopia", BaselineDE: 23.4967},
	{First: "ansi.cyan", Second: "ansi.white", Vision: "deuteranopia", BaselineDE: 30.3053},
	{First: "ansi.cyan", Second: "ansi.white", Vision: "tritanopia", BaselineDE: 31.9776},
	{First: "ansi.cyan", Second: "ansi.bright_black", Vision: "protanopia", BaselineDE: 47.3518},
	{First: "ansi.cyan", Second: "ansi.bright_black", Vision: "deuteranopia", BaselineDE: 47.3670},
	{First: "ansi.cyan", Second: "ansi.bright_black", Vision: "tritanopia", BaselineDE: 59.6517},
	{First: "ansi.cyan", Second: "ansi.bright_red", Vision: "protanopia", BaselineDE: 26.0303},
	{First: "ansi.cyan", Second: "ansi.bright_red", Vision: "deuteranopia", BaselineDE: 36.3783},
	{First: "ansi.cyan", Second: "ansi.bright_red", Vision: "tritanopia", BaselineDE: 80.3556},
	{First: "ansi.cyan", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 27.6529},
	{First: "ansi.cyan", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 27.1672},
	{First: "ansi.cyan", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 7.3086},
	{First: "ansi.cyan", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 54.9794},
	{First: "ansi.cyan", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 64.5390},
	{First: "ansi.cyan", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 65.1492},
	{First: "ansi.cyan", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 10.9742},
	{First: "ansi.cyan", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 16.1242},
	{First: "ansi.cyan", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 34.7262},
	{First: "ansi.cyan", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 3.9791},
	{First: "ansi.cyan", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 18.4190},
	{First: "ansi.cyan", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 58.2194},
	{First: "ansi.cyan", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 12.5599},
	{First: "ansi.cyan", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 16.3855},
	{First: "ansi.cyan", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 14.7592},
	{First: "ansi.cyan", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 33.0861},
	{First: "ansi.cyan", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 42.3645},
	{First: "ansi.cyan", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 46.5675},
	{First: "ansi.white", Second: "ansi.bright_black", Vision: "protanopia", BaselineDE: 52.9830},
	{First: "ansi.white", Second: "ansi.bright_black", Vision: "deuteranopia", BaselineDE: 51.3857},
	{First: "ansi.white", Second: "ansi.bright_black", Vision: "tritanopia", BaselineDE: 54.1019},
	{First: "ansi.white", Second: "ansi.bright_red", Vision: "protanopia", BaselineDE: 29.0601},
	{First: "ansi.white", Second: "ansi.bright_red", Vision: "deuteranopia", BaselineDE: 24.6471},
	{First: "ansi.white", Second: "ansi.bright_red", Vision: "tritanopia", BaselineDE: 55.9277},
	{First: "ansi.white", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 17.4729},
	{First: "ansi.white", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 19.2136},
	{First: "ansi.white", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 33.8039},
	{First: "ansi.white", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 37.4587},
	{First: "ansi.white", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 39.7089},
	{First: "ansi.white", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 37.8385},
	{First: "ansi.white", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 21.4590},
	{First: "ansi.white", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 20.2225},
	{First: "ansi.white", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 18.5062},
	{First: "ansi.white", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 19.5240},
	{First: "ansi.white", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 11.8884},
	{First: "ansi.white", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 29.7713},
	{First: "ansi.white", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 10.9393},
	{First: "ansi.white", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 13.9227},
	{First: "ansi.white", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 17.2392},
	{First: "ansi.white", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 9.8733},
	{First: "ansi.white", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 12.2931},
	{First: "ansi.white", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 14.7262},
	{First: "ansi.bright_black", Second: "ansi.bright_red", Vision: "protanopia", BaselineDE: 24.1775},
	{First: "ansi.bright_black", Second: "ansi.bright_red", Vision: "deuteranopia", BaselineDE: 28.7784},
	{First: "ansi.bright_black", Second: "ansi.bright_red", Vision: "tritanopia", BaselineDE: 43.6871},
	{First: "ansi.bright_black", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 37.0757},
	{First: "ansi.bright_black", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 32.2192},
	{First: "ansi.bright_black", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 55.9256},
	{First: "ansi.bright_black", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 49.7277},
	{First: "ansi.bright_black", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 52.5023},
	{First: "ansi.bright_black", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 44.6259},
	{First: "ansi.bright_black", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 37.6531},
	{First: "ansi.bright_black", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 36.8671},
	{First: "ansi.bright_black", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 35.8804},
	{First: "ansi.bright_black", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 47.7393},
	{First: "ansi.bright_black", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 47.5880},
	{First: "ansi.bright_black", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 46.7033},
	{First: "ansi.bright_black", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 49.2005},
	{First: "ansi.bright_black", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 47.3260},
	{First: "ansi.bright_black", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 55.4461},
	{First: "ansi.bright_black", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 60.1538},
	{First: "ansi.bright_black", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 59.7668},
	{First: "ansi.bright_black", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 59.8390},
	{First: "ansi.bright_red", Second: "ansi.bright_green", Vision: "protanopia", BaselineDE: 15.2183},
	{First: "ansi.bright_red", Second: "ansi.bright_green", Vision: "deuteranopia", BaselineDE: 9.4220},
	{First: "ansi.bright_red", Second: "ansi.bright_green", Vision: "tritanopia", BaselineDE: 79.0673},
	{First: "ansi.bright_red", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 38.3770},
	{First: "ansi.bright_red", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 31.2200},
	{First: "ansi.bright_red", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 20.1960},
	{First: "ansi.bright_red", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 15.1573},
	{First: "ansi.bright_red", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 20.3324},
	{First: "ansi.bright_red", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 45.6893},
	{First: "ansi.bright_red", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 25.2604},
	{First: "ansi.bright_red", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 25.8608},
	{First: "ansi.bright_red", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 28.1952},
	{First: "ansi.bright_red", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 25.2322},
	{First: "ansi.bright_red", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 26.7099},
	{First: "ansi.bright_red", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 68.6710},
	{First: "ansi.bright_red", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 37.0040},
	{First: "ansi.bright_red", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 31.2301},
	{First: "ansi.bright_red", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 50.0308},
	{First: "ansi.bright_green", Second: "ansi.bright_yellow", Vision: "protanopia", BaselineDE: 27.6537},
	{First: "ansi.bright_green", Second: "ansi.bright_yellow", Vision: "deuteranopia", BaselineDE: 38.4645},
	{First: "ansi.bright_green", Second: "ansi.bright_yellow", Vision: "tritanopia", BaselineDE: 64.2882},
	{First: "ansi.bright_green", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 18.5356},
	{First: "ansi.bright_green", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 11.0464},
	{First: "ansi.bright_green", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 34.1512},
	{First: "ansi.bright_green", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 24.8269},
	{First: "ansi.bright_green", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 17.1601},
	{First: "ansi.bright_green", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 58.6481},
	{First: "ansi.bright_green", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 19.6701},
	{First: "ansi.bright_green", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 17.7217},
	{First: "ansi.bright_green", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 17.5697},
	{First: "ansi.bright_green", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 23.1672},
	{First: "ansi.bright_green", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 28.8486},
	{First: "ansi.bright_green", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 48.3948},
	{First: "ansi.bright_yellow", Second: "ansi.bright_blue", Vision: "protanopia", BaselineDE: 46.1642},
	{First: "ansi.bright_yellow", Second: "ansi.bright_blue", Vision: "deuteranopia", BaselineDE: 48.8877},
	{First: "ansi.bright_yellow", Second: "ansi.bright_blue", Vision: "tritanopia", BaselineDE: 31.8934},
	{First: "ansi.bright_yellow", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 51.8005},
	{First: "ansi.bright_yellow", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 48.7149},
	{First: "ansi.bright_yellow", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 12.8641},
	{First: "ansi.bright_yellow", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 45.0412},
	{First: "ansi.bright_yellow", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 50.4427},
	{First: "ansi.bright_yellow", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 52.1794},
	{First: "ansi.bright_yellow", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 35.1483},
	{First: "ansi.bright_yellow", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 35.0308},
	{First: "ansi.bright_yellow", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 30.9215},
	{First: "ansi.bright_blue", Second: "ansi.bright_magenta", Vision: "protanopia", BaselineDE: 10.2165},
	{First: "ansi.bright_blue", Second: "ansi.bright_magenta", Vision: "deuteranopia", BaselineDE: 11.4574},
	{First: "ansi.bright_blue", Second: "ansi.bright_magenta", Vision: "tritanopia", BaselineDE: 25.8329},
	{First: "ansi.bright_blue", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 12.9523},
	{First: "ansi.bright_blue", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 10.6989},
	{First: "ansi.bright_blue", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 23.9580},
	{First: "ansi.bright_blue", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 31.1977},
	{First: "ansi.bright_blue", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 32.2076},
	{First: "ansi.bright_blue", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 26.8262},
	{First: "ansi.bright_magenta", Second: "ansi.bright_cyan", Vision: "protanopia", BaselineDE: 8.5852},
	{First: "ansi.bright_magenta", Second: "ansi.bright_cyan", Vision: "deuteranopia", BaselineDE: 2.0373},
	{First: "ansi.bright_magenta", Second: "ansi.bright_cyan", Vision: "tritanopia", BaselineDE: 44.7252},
	{First: "ansi.bright_magenta", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 29.1105},
	{First: "ansi.bright_magenta", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 24.0230},
	{First: "ansi.bright_magenta", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 22.7503},
	{First: "ansi.bright_cyan", Second: "ansi.bright_white", Vision: "protanopia", BaselineDE: 20.5809},
	{First: "ansi.bright_cyan", Second: "ansi.bright_white", Vision: "deuteranopia", BaselineDE: 26.0315},
	{First: "ansi.bright_cyan", Second: "ansi.bright_white", Vision: "tritanopia", BaselineDE: 31.8110},
}

var namedChecks = func() []check {
	checks := []check{
		{First: "semantic.error", Second: "semantic.warning", Vision: "protanopia", BaselineDE: 10.0475},
		{First: "semantic.error", Second: "semantic.warning", Vision: "deuteranopia", BaselineDE: 5.9503},
		{First: "semantic.error", Second: "semantic.warning", Vision: "tritanopia", BaselineDE: 15.7606},
		{First: "semantic.error", Second: "semantic.success", Vision: "protanopia", BaselineDE: 23.4639},
		{First: "semantic.error", Second: "semantic.success", Vision: "deuteranopia", BaselineDE: 12.1297},
		{First: "semantic.error", Second: "semantic.success", Vision: "tritanopia", BaselineDE: 53.2708},
		{First: "semantic.error", Second: "semantic.info", Vision: "protanopia", BaselineDE: 8.0931},
		{First: "semantic.error", Second: "semantic.info", Vision: "deuteranopia", BaselineDE: 3.9470},
		{First: "semantic.error", Second: "semantic.info", Vision: "tritanopia", BaselineDE: 44.2824},
		{First: "semantic.warning", Second: "semantic.success", Vision: "protanopia", BaselineDE: 13.5831},
		{First: "semantic.warning", Second: "semantic.success", Vision: "deuteranopia", BaselineDE: 6.8733},
		{First: "semantic.warning", Second: "semantic.success", Vision: "tritanopia", BaselineDE: 37.6345},
		{First: "semantic.warning", Second: "semantic.info", Vision: "protanopia", BaselineDE: 5.7424},
		{First: "semantic.warning", Second: "semantic.info", Vision: "deuteranopia", BaselineDE: 9.6291},
		{First: "semantic.warning", Second: "semantic.info", Vision: "tritanopia", BaselineDE: 28.6252},
		{First: "semantic.success", Second: "semantic.info", Vision: "protanopia", BaselineDE: 18.3617},
		{First: "semantic.success", Second: "semantic.info", Vision: "deuteranopia", BaselineDE: 16.0548},
		{First: "semantic.success", Second: "semantic.info", Vision: "tritanopia", BaselineDE: 10.3774},
		{First: "git.added", Second: "git.changed", Vision: "protanopia", BaselineDE: 23.4639},
		{First: "git.added", Second: "git.changed", Vision: "deuteranopia", BaselineDE: 12.1297},
		{First: "git.added", Second: "git.changed", Vision: "tritanopia", BaselineDE: 53.2708},
		{First: "git.added", Second: "git.deleted", Vision: "protanopia", BaselineDE: 7.8061},
		{First: "git.added", Second: "git.deleted", Vision: "deuteranopia", BaselineDE: 1.6964},
		{First: "git.added", Second: "git.deleted", Vision: "tritanopia", BaselineDE: 34.6446},
		{First: "git.changed", Second: "git.deleted", Vision: "protanopia", BaselineDE: 15.7665},
		{First: "git.changed", Second: "git.deleted", Vision: "deuteranopia", BaselineDE: 11.0439},
		{First: "git.changed", Second: "git.deleted", Vision: "tritanopia", BaselineDE: 18.6931},
		{First: "nvim.diff_add_bg", Second: "nvim.diff_change_bg", Vision: "protanopia", BaselineDE: 18.4806},
		{First: "nvim.diff_add_bg", Second: "nvim.diff_change_bg", Vision: "deuteranopia", BaselineDE: 15.7067},
		{First: "nvim.diff_add_bg", Second: "nvim.diff_change_bg", Vision: "tritanopia", BaselineDE: 13.4905},
		{First: "nvim.diff_add_bg", Second: "nvim.diff_delete_bg", Vision: "protanopia", BaselineDE: 8.9361},
		{First: "nvim.diff_add_bg", Second: "nvim.diff_delete_bg", Vision: "deuteranopia", BaselineDE: 7.7573},
		{First: "nvim.diff_add_bg", Second: "nvim.diff_delete_bg", Vision: "tritanopia", BaselineDE: 24.9067},
		{First: "nvim.diff_change_bg", Second: "nvim.diff_delete_bg", Vision: "protanopia", BaselineDE: 18.0261},
		{First: "nvim.diff_change_bg", Second: "nvim.diff_delete_bg", Vision: "deuteranopia", BaselineDE: 19.7908},
		{First: "nvim.diff_change_bg", Second: "nvim.diff_delete_bg", Vision: "tritanopia", BaselineDE: 13.3098},
		{First: "nvim.diff_add_inline_bg", Second: "nvim.diff_change_inline_bg", Vision: "protanopia", BaselineDE: 24.2167},
		{First: "nvim.diff_add_inline_bg", Second: "nvim.diff_change_inline_bg", Vision: "deuteranopia", BaselineDE: 18.6511},
		{First: "nvim.diff_add_inline_bg", Second: "nvim.diff_change_inline_bg", Vision: "tritanopia", BaselineDE: 25.4855},
		{First: "nvim.diff_add_inline_bg", Second: "nvim.diff_delete_inline_bg", Vision: "protanopia", BaselineDE: 14.0986},
		{First: "nvim.diff_add_inline_bg", Second: "nvim.diff_delete_inline_bg", Vision: "deuteranopia", BaselineDE: 15.6908},
		{First: "nvim.diff_add_inline_bg", Second: "nvim.diff_delete_inline_bg", Vision: "tritanopia", BaselineDE: 50.7671},
		{First: "nvim.diff_change_inline_bg", Second: "nvim.diff_delete_inline_bg", Vision: "protanopia", BaselineDE: 26.8175},
		{First: "nvim.diff_change_inline_bg", Second: "nvim.diff_delete_inline_bg", Vision: "deuteranopia", BaselineDE: 31.4144},
		{First: "nvim.diff_change_inline_bg", Second: "nvim.diff_delete_inline_bg", Vision: "tritanopia", BaselineDE: 27.5152},
		{First: "semantic.keyword", Second: "semantic.string", Vision: "protanopia", BaselineDE: 25.4922},
		{First: "semantic.keyword", Second: "semantic.string", Vision: "deuteranopia", BaselineDE: 22.4409},
		{First: "semantic.keyword", Second: "semantic.string", Vision: "tritanopia", BaselineDE: 18.8108},
		{First: "semantic.keyword", Second: "semantic.function", Vision: "protanopia", BaselineDE: 11.4017},
		{First: "semantic.keyword", Second: "semantic.function", Vision: "deuteranopia", BaselineDE: 2.4413},
		{First: "semantic.keyword", Second: "semantic.function", Vision: "tritanopia", BaselineDE: 42.0782},
		{First: "semantic.keyword", Second: "semantic.type", Vision: "protanopia", BaselineDE: 20.8251},
		{First: "semantic.keyword", Second: "semantic.type", Vision: "deuteranopia", BaselineDE: 20.1942},
		{First: "semantic.keyword", Second: "semantic.type", Vision: "tritanopia", BaselineDE: 8.3706},
		{First: "semantic.keyword", Second: "semantic.constant", Vision: "protanopia", BaselineDE: 26.2347},
		{First: "semantic.keyword", Second: "semantic.constant", Vision: "deuteranopia", BaselineDE: 26.0247},
		{First: "semantic.keyword", Second: "semantic.constant", Vision: "tritanopia", BaselineDE: 12.1247},
		{First: "semantic.function", Second: "semantic.string", Vision: "protanopia", BaselineDE: 15.6132},
		{First: "semantic.function", Second: "semantic.string", Vision: "deuteranopia", BaselineDE: 20.0202},
		{First: "semantic.function", Second: "semantic.string", Vision: "tritanopia", BaselineDE: 24.3265},
		{First: "semantic.function", Second: "semantic.type", Vision: "protanopia", BaselineDE: 11.4060},
		{First: "semantic.function", Second: "semantic.type", Vision: "deuteranopia", BaselineDE: 17.7817},
		{First: "semantic.function", Second: "semantic.type", Vision: "tritanopia", BaselineDE: 35.4172},
		{First: "semantic.string", Second: "semantic.type", Vision: "protanopia", BaselineDE: 4.6974},
		{First: "semantic.string", Second: "semantic.type", Vision: "deuteranopia", BaselineDE: 2.2951},
		{First: "semantic.string", Second: "semantic.type", Vision: "tritanopia", BaselineDE: 11.2748},
		{First: "semantic.constant", Second: "semantic.string", Vision: "protanopia", BaselineDE: 9.4729},
		{First: "semantic.constant", Second: "semantic.string", Vision: "deuteranopia", BaselineDE: 11.0755},
		{First: "semantic.constant", Second: "semantic.string", Vision: "tritanopia", BaselineDE: 17.7916},
		{First: "semantic.constant", Second: "semantic.type", Vision: "protanopia", BaselineDE: 10.5573},
		{First: "semantic.constant", Second: "semantic.type", Vision: "deuteranopia", BaselineDE: 11.1040},
		{First: "semantic.constant", Second: "semantic.type", Vision: "tritanopia", BaselineDE: 10.7933},
		{First: "semantic.variable", Second: "semantic.constant", Vision: "protanopia", BaselineDE: 8.9508},
		{First: "semantic.variable", Second: "semantic.constant", Vision: "deuteranopia", BaselineDE: 6.6581},
		{First: "semantic.variable", Second: "semantic.constant", Vision: "tritanopia", BaselineDE: 10.7056},
		{First: "semantic.builtin_variable", Second: "semantic.keyword", Vision: "protanopia", BaselineDE: 19.9496},
		{First: "semantic.builtin_variable", Second: "semantic.keyword", Vision: "deuteranopia", BaselineDE: 14.0765},
		{First: "semantic.builtin_variable", Second: "semantic.keyword", Vision: "tritanopia", BaselineDE: 32.5324},
		{First: "wezterm.active_tab", Second: "wezterm.inactive_tab", Vision: "protanopia", BaselineDE: 20.4237},
		{First: "wezterm.active_tab", Second: "wezterm.inactive_tab", Vision: "deuteranopia", BaselineDE: 21.0455},
		{First: "wezterm.active_tab", Second: "wezterm.inactive_tab", Vision: "tritanopia", BaselineDE: 12.8677},
		{First: "semantic.variable", Second: "semantic.string", Vision: "protanopia", BaselineDE: 16.1447},
		{First: "semantic.variable", Second: "semantic.string", Vision: "deuteranopia", BaselineDE: 16.9574},
		{First: "semantic.variable", Second: "semantic.string", Vision: "tritanopia", BaselineDE: 16.4429},
		{First: "semantic.variable", Second: "semantic.type", Vision: "protanopia", BaselineDE: 18.7940},
		{First: "semantic.variable", Second: "semantic.type", Vision: "deuteranopia", BaselineDE: 17.4195},
		{First: "semantic.variable", Second: "semantic.type", Vision: "tritanopia", BaselineDE: 16.7355},
		{First: "semantic.error", Second: "git.changed", Vision: "protanopia", BaselineDE: 20.0, Kind: KindAchievedFreeze},
		{First: "semantic.error", Second: "git.changed", Vision: "deuteranopia", BaselineDE: 20.0, Kind: KindAchievedFreeze},
		{First: "semantic.error", Second: "git.changed", Vision: "tritanopia", BaselineDE: 18.7517, Kind: KindAchievedFreeze},
		{First: "foregrounds.heading", Second: "semantic.info", Vision: "protanopia", BaselineDE: 3.0, Kind: KindMinimumSeparation},
		{First: "foregrounds.heading", Second: "semantic.info", Vision: "deuteranopia", BaselineDE: 3.0, Kind: KindMinimumSeparation},
		{First: "foregrounds.heading", Second: "semantic.info", Vision: "tritanopia", BaselineDE: 3.0, Kind: KindMinimumSeparation},
		{First: "foregrounds.heading", Second: "semantic.builtin_variable", Vision: "protanopia", BaselineDE: 3.0, Kind: KindMinimumSeparation},
		{First: "foregrounds.heading", Second: "semantic.builtin_variable", Vision: "deuteranopia", BaselineDE: 3.0, Kind: KindMinimumSeparation},
		{First: "foregrounds.heading", Second: "semantic.builtin_variable", Vision: "tritanopia", BaselineDE: 3.0, Kind: KindMinimumSeparation},
		{First: "semantic.info", Second: "semantic.builtin_variable", Vision: "protanopia", BaselineDE: 3.0, Kind: KindMinimumSeparation},
		{First: "semantic.info", Second: "semantic.builtin_variable", Vision: "deuteranopia", BaselineDE: 3.0, Kind: KindMinimumSeparation},
		{First: "semantic.info", Second: "semantic.builtin_variable", Vision: "tritanopia", BaselineDE: 3.0, Kind: KindMinimumSeparation},
	}
	for index := range checks[:87] {
		checks[index].Kind = KindHistoricalFloor
	}
	return checks
}()

var waivers = []waiver{
	{First: "ansi.bright_blue", Second: "ansi.white", Vision: "protanopia", ApprovedDE: 18.7940},
	{First: "ansi.bright_blue", Second: "ansi.white", Vision: "deuteranopia", ApprovedDE: 17.4195},
	{First: "ansi.bright_green", Second: "ansi.bright_white", Vision: "protanopia", ApprovedDE: 17.6550},
}

var namedPairOrder = [][2]string{
	{"semantic.error", "semantic.warning"},
	{"semantic.error", "semantic.success"},
	{"semantic.error", "semantic.info"},
	{"semantic.warning", "semantic.success"},
	{"semantic.warning", "semantic.info"},
	{"semantic.success", "semantic.info"},
	{"git.added", "git.changed"},
	{"git.added", "git.deleted"},
	{"git.changed", "git.deleted"},
	{"nvim.diff_add_bg", "nvim.diff_change_bg"},
	{"nvim.diff_add_bg", "nvim.diff_delete_bg"},
	{"nvim.diff_change_bg", "nvim.diff_delete_bg"},
	{"nvim.diff_add_inline_bg", "nvim.diff_change_inline_bg"},
	{"nvim.diff_add_inline_bg", "nvim.diff_delete_inline_bg"},
	{"nvim.diff_change_inline_bg", "nvim.diff_delete_inline_bg"},
	{"semantic.keyword", "semantic.string"},
	{"semantic.keyword", "semantic.function"},
	{"semantic.keyword", "semantic.type"},
	{"semantic.keyword", "semantic.constant"},
	{"semantic.function", "semantic.string"},
	{"semantic.function", "semantic.type"},
	{"semantic.string", "semantic.type"},
	{"semantic.constant", "semantic.string"},
	{"semantic.constant", "semantic.type"},
	{"semantic.variable", "semantic.constant"},
	{"semantic.builtin_variable", "semantic.keyword"},
	{"wezterm.active_tab", "wezterm.inactive_tab"},
	{"semantic.variable", "semantic.string"},
	{"semantic.variable", "semantic.type"},
}

var namedPairAllowlists = map[checkKind][][2]string{
	KindHistoricalFloor: namedPairOrder,
	KindAchievedFreeze: {
		{"semantic.error", "git.changed"},
	},
	KindMinimumSeparation: {
		{"foregrounds.heading", "semantic.info"},
		{"foregrounds.heading", "semantic.builtin_variable"},
		{"semantic.info", "semantic.builtin_variable"},
	},
}

var ansiNames = []string{
	"ansi.black", "ansi.red", "ansi.green", "ansi.yellow",
	"ansi.blue", "ansi.magenta", "ansi.cyan", "ansi.white",
	"ansi.bright_black", "ansi.bright_red", "ansi.bright_green", "ansi.bright_yellow",
	"ansi.bright_blue", "ansi.bright_magenta", "ansi.bright_cyan", "ansi.bright_white",
}

var visions = []string{"protanopia", "deuteranopia", "tritanopia"}

func main() {
	palettePath, err := palette.ResolveDefaultPath(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify-cvd-pairs: %v\n", err)
		os.Exit(1)
	}
	colorPalette, err := verifycolors.LoadPalette(palettePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "verify-cvd-pairs: load palette: %v\n", err)
		os.Exit(1)
	}
	values := colorPalette.TokenValues()
	if err := validateTables(values); err != nil {
		fmt.Fprintf(os.Stderr, "verify-cvd-pairs: invalid tables: %v\n", err)
		os.Exit(1)
	}
	if err := evaluate(values); err != nil {
		fmt.Fprintf(os.Stderr, "verify-cvd-pairs: %v\n", err)
		os.Exit(1)
	}
	historicalPairs, historicalRows := namedKindCounts(KindHistoricalFloor)
	achievedPairs, achievedRows := namedKindCounts(KindAchievedFreeze)
	minimumPairs, minimumRows := namedKindCounts(KindMinimumSeparation)
	fmt.Printf("ansiChecks=%d namedChecks=%d waivers=%d\n", len(ansiChecks), len(namedChecks), len(waivers))
	fmt.Printf("namedChecks total=%d pairs/%d rows\n", historicalPairs+achievedPairs+minimumPairs, historicalRows+achievedRows+minimumRows)
	fmt.Printf("namedChecks kinds: historical=%d/%d achieved=%d/%d\n", historicalPairs, historicalRows, achievedPairs, achievedRows)
	fmt.Printf("namedChecks minimum separation: %d/%d\n", minimumPairs, minimumRows)
	fmt.Println("CVD pair verification passed.")
}

func validateTables(values map[verifycolors.TokenRef]string) error {
	root, err := palette.FindRepositoryRoot()
	if err != nil {
		return fmt.Errorf("find repository root: %w", err)
	}
	fixturePath := filepath.Join(root, "scripts", "testdata", "legacy-palette-8bf163c.toml")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		return fmt.Errorf("read legacy palette fixture: %w", err)
	}
	legacyPalette, err := verifycolors.LoadPalette(fixturePath)
	if err != nil {
		return fmt.Errorf("load legacy palette fixture: %w", err)
	}
	return validateTablesWithFixture(values, fixture, legacyPalette.TokenValues())
}

func validateTablesWithFixture(values map[verifycolors.TokenRef]string, fixture []byte, legacyValues map[verifycolors.TokenRef]string) error {
	digest := sha256.Sum256(fixture)
	if got := hex.EncodeToString(digest[:]); got != legacyFixtureSHA256 {
		return fmt.Errorf("legacy palette fixture sha256 = %s, want %s", got, legacyFixtureSHA256)
	}
	if len(ansiChecks) != 360 {
		return fmt.Errorf("ansiChecks count = %d, want 360", len(ansiChecks))
	}
	if len(namedChecks) != 99 {
		return fmt.Errorf("namedChecks count = %d, want 99", len(namedChecks))
	}
	if len(waivers) != 3 {
		return fmt.Errorf("waivers count = %d, want 3", len(waivers))
	}
	expectedANSI := make(map[string]bool, 360)
	for first := 0; first < len(ansiNames); first++ {
		for second := first + 1; second < len(ansiNames); second++ {
			for _, vision := range visions {
				expectedANSI[triple(ansiNames[first], ansiNames[second], vision)] = true
			}
		}
	}
	seenANSI := make(map[string]bool, len(ansiChecks))
	baselineBelow, baselineAbove := 0, 0
	for _, item := range ansiChecks {
		key := triple(item.First, item.Second, item.Vision)
		if seenANSI[key] {
			return fmt.Errorf("ansiChecks duplicate %s", key)
		}
		seenANSI[key] = true
		if !expectedANSI[key] {
			return fmt.Errorf("ansiChecks unexpected triple %s", key)
		}
		if item.BaselineDE < cvdThreshold {
			baselineBelow++
		} else {
			baselineAbove++
		}
		if _, ok := values[verifycolors.TokenRef(item.First)]; !ok {
			return fmt.Errorf("ansiChecks token %q is missing from palette", item.First)
		}
		if _, ok := values[verifycolors.TokenRef(item.Second)]; !ok {
			return fmt.Errorf("ansiChecks token %q is missing from palette", item.Second)
		}
	}
	if len(seenANSI) != len(expectedANSI) {
		return fmt.Errorf("ansiChecks universe has %d triples, want %d", len(seenANSI), len(expectedANSI))
	}
	if baselineBelow != 67 || baselineAbove != 293 {
		return fmt.Errorf("ansiChecks baseline split = %d/%d, want 67/293", baselineBelow, baselineAbove)
	}
	seenNamed := make(map[string]int, len(namedChecks))
	seenNamedPairs := make(map[checkKind]map[string]bool)
	kindRows := make(map[checkKind]int)
	for kind := range namedPairAllowlists {
		seenNamedPairs[kind] = make(map[string]bool)
	}
	for _, item := range namedChecks {
		key := triple(item.First, item.Second, item.Vision)
		seenNamed[key]++
		kindRows[item.Kind]++
		pair := pairKey(item.First, item.Second)
		if _, ok := namedPairAllowlists[item.Kind]; !ok {
			return fmt.Errorf("namedChecks has unknown kind %q", item.Kind)
		}
		allowed := false
		for _, candidate := range namedPairAllowlists[item.Kind] {
			if pair == pairKey(candidate[0], candidate[1]) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("namedChecks pair %s has unexpected kind %q", pair, item.Kind)
		}
		seenNamedPairs[item.Kind][pair] = true
		if _, ok := values[verifycolors.TokenRef(item.First)]; !ok {
			return fmt.Errorf("namedChecks token %q is missing from palette", item.First)
		}
		if _, ok := values[verifycolors.TokenRef(item.Second)]; !ok {
			return fmt.Errorf("namedChecks token %q is missing from palette", item.Second)
		}
		switch item.Kind {
		case KindHistoricalFloor:
			legacyDE, err := deltaFor(legacyValues, item)
			if err != nil {
				return fmt.Errorf("historical %s: %w", key, err)
			}
			if item.BaselineDE != floor4(legacyDE) {
				return fmt.Errorf("historical %s baseline = %.4f, want floor4(legacy dE %.9f) = %.4f", key, item.BaselineDE, legacyDE, floor4(legacyDE))
			}
		case KindAchievedFreeze:
			want := map[string]float64{"protanopia": 20.0, "deuteranopia": 20.0, "tritanopia": 18.7517}[item.Vision]
			if item.BaselineDE != want {
				return fmt.Errorf("achieved %s baseline = %.4f, want %.4f", key, item.BaselineDE, want)
			}
		case KindMinimumSeparation:
			if item.BaselineDE != 3.0 {
				return fmt.Errorf("minimum separation %s baseline = %.4f, want 3.0000", key, item.BaselineDE)
			}
			legacyDE, err := deltaFor(legacyValues, item)
			if err != nil {
				return fmt.Errorf("minimum separation %s: %w", key, err)
			}
			if legacyDE != 0 {
				return fmt.Errorf("minimum separation %s legacy dE = %.9f, want 0", key, legacyDE)
			}
		}
	}
	if len(seenNamed) != 99 {
		return fmt.Errorf("namedChecks has %d triples, want 99", len(seenNamed))
	}
	for kind, allowlist := range namedPairAllowlists {
		for _, pair := range allowlist {
			for _, vision := range visions {
				if seenNamed[triple(pair[0], pair[1], vision)] != 1 {
					return fmt.Errorf("namedChecks pair %s/%s vision %s is not present exactly once", pair[0], pair[1], vision)
				}
			}
		}
		if len(seenNamedPairs[kind]) != len(allowlist) {
			return fmt.Errorf("namedChecks kind %q has %d pairs, want %d", kind, len(seenNamedPairs[kind]), len(allowlist))
		}
	}
	wantRows := map[checkKind]int{KindHistoricalFloor: 87, KindAchievedFreeze: 3, KindMinimumSeparation: 9}
	for kind, want := range wantRows {
		if kindRows[kind] != want {
			return fmt.Errorf("namedChecks kind %q has %d rows, want %d", kind, kindRows[kind], want)
		}
	}
	for _, item := range waivers {
		key := triple(item.First, item.Second, item.Vision)
		found := false
		for _, ansi := range ansiChecks {
			if triple(ansi.First, ansi.Second, ansi.Vision) == key {
				if ansi.BaselineDE < cvdThreshold {
					return fmt.Errorf("waiver %s is not on baseline >= 20 side", key)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("waiver %s is not in ansiChecks", key)
		}
		if _, ok := values[verifycolors.TokenRef(item.First)]; !ok {
			return fmt.Errorf("waiver token %q is missing from palette", item.First)
		}
		if _, ok := values[verifycolors.TokenRef(item.Second)]; !ok {
			return fmt.Errorf("waiver token %q is missing from palette", item.Second)
		}
	}
	return nil
}

func evaluate(values map[verifycolors.TokenRef]string) error {
	below := 0
	for _, item := range ansiChecks {
		current, err := deltaFor(values, item)
		if err != nil {
			return err
		}
		if current < cvdThreshold {
			below++
		}
	}
	if below > 67 {
		return fmt.Errorf("criterion A failed: ansi dE < 20 count = %d, want <= 67", below)
	}
	for _, item := range ansiChecks {
		current, err := deltaFor(values, item)
		if err != nil {
			return err
		}
		if item.BaselineDE < cvdThreshold {
			continue
		}
		if waiver := findWaiver(item); waiver != nil {
			if current >= cvdThreshold {
				return fmt.Errorf("stale waiver %s = %.6f", triple(item.First, item.Second, item.Vision), current)
			}
			if current < waiver.ApprovedDE {
				return fmt.Errorf("waiver %s = %.6f, want >= %.4f", triple(item.First, item.Second, item.Vision), current, waiver.ApprovedDE)
			}
			continue
		}
		if current < cvdThreshold {
			return fmt.Errorf("criterion B failed: %s = %.6f, want >= 20", triple(item.First, item.Second, item.Vision), current)
		}
	}
	// namedChecks stores raw historical/achieved floors; evaluation applies the
	// cvdThreshold cap separately, while minimum-separation rows keep their 3.0 floor.
	for _, item := range namedChecks {
		current, err := deltaFor(values, item)
		if err != nil {
			return err
		}
		if item.Kind == KindMinimumSeparation {
			if current < item.BaselineDE {
				return fmt.Errorf("criterion C minimum separation failed: %s = %.6f, want >= %.4f", triple(item.First, item.Second, item.Vision), current, item.BaselineDE)
			}
			continue
		}
		limit := effectiveNamedLimit(item.BaselineDE)
		if current < limit {
			return fmt.Errorf("criterion C failed: %s = %.6f, want >= %.4f", triple(item.First, item.Second, item.Vision), current, limit)
		}
	}
	return nil
}

func deltaFor(values map[verifycolors.TokenRef]string, item check) (float64, error) {
	first := values[verifycolors.TokenRef(item.First)]
	second := values[verifycolors.TokenRef(item.Second)]
	for _, kind := range cvd.Types {
		if kind.Name != item.Vision {
			continue
		}
		firstLab, err := cvd.Apply(first, kind.Matrix)
		if err != nil {
			return 0, fmt.Errorf("%s first color: %w", triple(item.First, item.Second, item.Vision), err)
		}
		secondLab, err := cvd.Apply(second, kind.Matrix)
		if err != nil {
			return 0, fmt.Errorf("%s second color: %w", triple(item.First, item.Second, item.Vision), err)
		}
		return cvd.DeltaE76(firstLab, secondLab), nil
	}
	return 0, fmt.Errorf("unknown vision type %q", item.Vision)
}

func findWaiver(item check) *waiver {
	for index := range waivers {
		if triple(item.First, item.Second, item.Vision) == triple(waivers[index].First, waivers[index].Second, waivers[index].Vision) {
			return &waivers[index]
		}
	}
	return nil
}

func triple(first, second, vision string) string {
	if first > second {
		first, second = second, first
	}
	return first + "|" + second + "|" + vision
}

func pairKey(first, second string) string {
	if first > second {
		first, second = second, first
	}
	return first + "|" + second
}

func floor4(value float64) float64 {
	return math.Floor(value*10000) / 10000
}

func effectiveNamedLimit(recorded float64) float64 {
	if recorded > cvdThreshold {
		return cvdThreshold
	}
	return recorded
}

func namedKindCounts(kind checkKind) (pairs, rows int) {
	seen := make(map[string]bool)
	for _, item := range namedChecks {
		if item.Kind != kind {
			continue
		}
		rows++
		seen[pairKey(item.First, item.Second)] = true
	}
	return len(seen), rows
}
