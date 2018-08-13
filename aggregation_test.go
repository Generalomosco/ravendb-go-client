package ravendb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func NewOrders_All() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("Orders_All")
	res.Map = "docs.AggOrders.Select(order => new { order.currency,\n" +
		"                          order.product,\n" +
		"                          order.total,\n" +
		"                          order.quantity,\n" +
		"                          order.region,\n" +
		"                          order.at,\n" +
		"                          order.tax })"
	return res
}

type Currency = string

// Note: must rename as it conflicts with Order in order_test.go
type AggOrder struct {
	Currency Currency `json:"currency"`
	Product  string   `json:"product"`
	Total    float64  `json:"total"`
	Region   int      `json:"region"`
}

const (
	EUR = "EUR"
	PLN = "PLN"
	NIS = "NIS"
)

func aggregation_canCorrectlyAggregate_Double(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		obj := &AggOrder{
			Currency: EUR,
			Product:  "Milk",
			Total:    1.1,
			Region:   1,
		}

		obj2 := &AggOrder{
			Currency: EUR,
			Product:  "Milk",
			Total:    1,
			Region:   1,
		}
		err = session.Store(obj)
		assert.NoError(t, err)
		err = session.Store(obj2)
		assert.NoError(t, err)

		err = session.SaveChanges()

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.queryInIndex(getTypeOf(&Order{}), index)
		builder := func(f IFacetBuilder) {
			f.byField("region").maxOn("total").minOn("total")
		}
		q2 := q.aggregateBy(builder)
		result, err := q2.execute()
		assert.NoError(t, err)

		facetResult := result["region"]

		values := facetResult.getValues()
		val := values[0]
		assert.Equal(t, val.getCount(), 2)
		assert.Equal(t, *val.getMin(), float64(1))
		assert.Equal(t, *val.getMax(), float64(1.1))

		n := 0
		for _, x := range values {
			if x.getRange() == "1" {
				n++
			}
		}
		assert.Equal(t, n, 1)

		session.Close()
	}
}

func aggregation_canCorrectlyAggregate_MultipleItems(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewOrders_All()
	err = index.execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		obj := &AggOrder{
			Currency: EUR,
			Product:  "Milk",
			Total:    3,
		}

		obj2 := &AggOrder{
			Currency: NIS,
			Product:  "Milk",
			Total:    9,
		}

		obj3 := &AggOrder{
			Currency: EUR,
			Product:  "iPhone",
			Total:    3333,
		}

		err = session.Store(obj)
		assert.NoError(t, err)
		err = session.Store(obj2)
		assert.NoError(t, err)
		err = session.Store(obj3)
		assert.NoError(t, err)

		err = session.SaveChanges()

		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		q := session.queryInIndex(getTypeOf(&AggOrder{}), index)
		builder := func(f IFacetBuilder) {
			f.byField("product").sumOn("total")
		}
		q2 := q.aggregateBy(builder)
		builder2 := func(f IFacetBuilder) {
			f.byField("currency").sumOn("total")
		}
		q2 = q2.andAggregateBy(builder2)
		r, err := q2.execute()
		assert.NoError(t, err)

		facetResult := r["product"]

		values := facetResult.getValues()
		assert.Equal(t, len(values), 2)

		var n float64 = -1
		for _, x := range values {
			if x.getRange() == "milk" {
				n = *x.getSum()
				break
			}
		}
		assert.Equal(t, n, float64(12))

		n = -1
		for _, x := range values {
			if x.getRange() == "iphone" {
				n = *x.getSum()
				break
			}
		}
		assert.Equal(t, n, float64(3333))

		facetResult = r["currency"]
		values = facetResult.getValues()
		assert.Equal(t, len(values), 2)

		n = -1
		for _, x := range values {
			if x.getRange() == "eur" {
				n = *x.getSum()
				break
			}
		}
		assert.Equal(t, n, float64(3336))

		n = -1
		for _, x := range values {
			if x.getRange() == "nis" {
				n = *x.getSum()
				break
			}
		}
		assert.Equal(t, n, float64(9))

		session.Close()
	}
}
func aggregation_canCorrectlyAggregate_MultipleAggregations(t *testing.T)             {}
func aggregation_canCorrectlyAggregate_DisplayName(t *testing.T)                      {}
func aggregation_canCorrectlyAggregate_Ranges(t *testing.T)                           {}
func aggregation_canCorrectlyAggregate_DateTimeDataType_WithRangeCounts(t *testing.T) {}

type ItemsOrder struct {
	Items []string  `json:"items"`
	At    time.Time `json:"at"`
}

func NewItemsOrders_All() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("ItemsOrders_All")
	res.Map = "docs.ItemsOrders.Select(order => new { order.at,\n" +
		"                          order.items })"
	return res
}

func TestAggregation(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	aggregation_canCorrectlyAggregate_Double(t)
	aggregation_canCorrectlyAggregate_Ranges(t)
	aggregation_canCorrectlyAggregate_MultipleItems(t)
	aggregation_canCorrectlyAggregate_MultipleAggregations(t)
	aggregation_canCorrectlyAggregate_DateTimeDataType_WithRangeCounts(t)
	aggregation_canCorrectlyAggregate_DisplayName(t)
}
