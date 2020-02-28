package testing

import (
	"fmt"
	"github.com/JanFant/TLServer/data"
	"github.com/joho/godotenv"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPoint_StrToFloat(t *testing.T) {
	tables := []struct {
		str   string
		point data.Point
	}{
		{str: "(-30.45,40.59)", point: data.Point{Y: -30.45, X: 40.59}},
		{str: "( 40.3 , 50.8 )", point: data.Point{Y: 40.3, X: 50.8}},
		{str: "()", point: data.Point{Y: 0, X: 0}},
		{str: "(,)", point: data.Point{Y: 0, X: 0}},
	}
	convey.Convey("Convert str to Point", t, func() {
		var testPoint data.Point
		for _, table := range tables {
			testPoint.StrToFloat(table.str)
			convey.Convey(fmt.Sprintf("str: %v", table.str), func() {
				convey.So(testPoint, convey.ShouldResemble, table.point)
			})
		}
	})
}

func TestPoint_GetPoint(t *testing.T) {
	tables := []struct {
		point data.Point
		test  struct {
			x, y float64
		}
	}{
		{point: data.Point{Y: 23.1, X: 32.1}},
		{point: data.Point{Y: -14.71, X: 51.52}},
	}
	convey.Convey("GetPoint", t, func() {
		for _, table := range tables {
			convey.Convey(fmt.Sprintf("getData: %v", table.point), func() {
				table.test.y, table.test.x = table.point.GetPoint()
				convey.So(table.test.x, convey.ShouldEqual, table.point.X)
				convey.So(table.test.y, convey.ShouldEqual, table.point.Y)
			})
		}
	})
}

func TestPoint_SetPoint(t *testing.T) {
	tables := []struct {
		point data.Point
		test  struct {
			x, y float64
		}
	}{
		{test: struct{ x, y float64 }{x: 23.1, y: 32.1}},
		{test: struct{ x, y float64 }{x: -14.71, y: 52.51}},
	}
	convey.Convey("SetPoint", t, func() {
		for _, table := range tables {
			convey.Convey(fmt.Sprintf("setData : %v", table.test), func() {
				table.point.SetPoint(table.test.y, table.test.x)
				convey.So(table.test.x, convey.ShouldEqual, table.point.X)
				convey.So(table.test.y, convey.ShouldEqual, table.point.Y)

			})
		}
	})
}

func TestTakePointFromBD(t *testing.T) {
	_ = godotenv.Load("../.env")
	_ = data.ConnectDB()
	tables := []struct {
		info struct {
			region, area, id string
		}
		point data.Point
	}{
		{info: struct{ region, area, id string }{region: "2", area: "3", id: "52"}, point: data.Point{Y: 66.86672758094775, X: -172.0455717109074}},
		{info: struct{ region, area, id string }{region: "2", area: "3", id: "53"}, point: data.Point{Y: 66.86672758094775, X: -173.0455717109074}},
		{info: struct{ region, area, id string }{region: "2", area: "3", id: "51"}, point: data.Point{Y: 66.86672758094775, X: -171.0455717109074}},
		{info: struct{ region, area, id string }{region: "1000", area: "1000", id: "1000"}},
		{info: struct{ region, area, id string }{region: "", area: "", id: ""},},
	}
	convey.Convey("Correct location on the map", t, func() {
		for _, table := range tables[:3] {
			convey.Convey(fmt.Sprintf("Location region: %v, area: %v, id: %v", table.info.region, table.info.area, table.info.id), func() {
				point, err := data.TakePointFromBD(table.info.region, table.info.area, table.info.id)
				convey.So(point.X, convey.ShouldNotBeEmpty)
				convey.So(point.Y, convey.ShouldNotBeEmpty)
				convey.So(point.X, convey.ShouldEqual, table.point.X)
				convey.So(point.Y, convey.ShouldEqual, table.point.Y)
				convey.So(err, convey.ShouldEqual, nil)
			})
		}
	})
	convey.Convey("Nonexistent map location", t, func() {
		table := tables[3]
		convey.Convey(fmt.Sprintf("Location region: %v, area: %v, id: %v", table.info.region, table.info.area, table.info.id), func() {
			point, err := data.TakePointFromBD(table.info.region, table.info.area, table.info.id)
			convey.So(point.X, convey.ShouldEqual, 0)
			convey.So(point.Y, convey.ShouldEqual, 0)
			convey.So(err, convey.ShouldBeError)
		})
	})
	convey.Convey("Unfilled map location", t, func() {
		table := tables[4]
		convey.Convey(fmt.Sprintf("Location region: %v, area: %v, id: %v", table.info.region, table.info.area, table.info.id), func() {
			point, err := data.TakePointFromBD(table.info.region, table.info.area, table.info.id)
			convey.So(point.X, convey.ShouldEqual, 0)
			convey.So(point.Y, convey.ShouldEqual, 0)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

