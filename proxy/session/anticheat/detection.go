/*
 *
 *           ____                        ______
 *          / __ \_________  _  ____  __/ ____/_  _____
 *         / /_/ / ___/ __ \| |/_/ / / / __/ / / / / _ \
 *        / ____/ /  / /_/ />  </ /_/ / /___/ /_/ /  __/
 *       /_/   /_/   \____/_/|_|\__, /_____/\__, /\___/
 *                                /_/         /_/
 *       ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀
 *
 *                     Author: Suremeo (github.com/Suremeo)
 *
 *
 */

package anticheat

type Detection struct { // bson is to make it look better when storing in MongoDB :)
	Type      DetectionType          `bson:"Type"`
	Arguments map[string]interface{} `bson:"Arguments"`
}

type DetectionType string

const (
	DetectionTypeFly         = "FLY"
	DetectionTypeSpeed       = "SPEED"
	DetectionTypeKillaura    = "KILLAURA"
	DetectionTypeTimer       = "TIMER"
	DetectionTypeAutoclicker = "AUTOCLICKER"
)
