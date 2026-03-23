package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/vasudevchavan/kubecaps/pkg/types"

	prometheusapi "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Client wraps the Prometheus HTTP API client.
type Client struct {
	api v1.API
	url string
}

// NewClient creates a new Prometheus API client.
func NewClient(url string) (*Client, error) {
	client, err := prometheusapi.NewClient(prometheusapi.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus client: %w", err)
	}

	return &Client{
		api: v1.NewAPI(client),
		url: url,
	}, nil
}

// QueryRange executes a range query and returns time-series data points.
func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]types.DataPoint, error) {
	result, warnings, err := c.api.QueryRange(ctx, query, v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		return nil, fmt.Errorf("prometheus range query failed: %w", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("⚠️  Prometheus warnings: %v\n", warnings)
	}

	return extractDataPoints(result)
}

// QueryInstant executes an instant query and returns the current value.
func (c *Client) QueryInstant(ctx context.Context, query string) (float64, error) {
	result, warnings, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("prometheus query failed: %w", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("⚠️  Prometheus warnings: %v\n", warnings)
	}

	return extractScalar(result)
}

// QueryRangeMulti executes a range query returning multiple series (keyed by label).
func (c *Client) QueryRangeMulti(ctx context.Context, query string, start, end time.Time, step time.Duration) (map[string][]types.DataPoint, error) {
	result, warnings, err := c.api.QueryRange(ctx, query, v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		return nil, fmt.Errorf("prometheus range query failed: %w", err)
	}
	if len(warnings) > 0 {
		fmt.Printf("⚠️  Prometheus warnings: %v\n", warnings)
	}

	return extractMultiSeries(result)
}

// extractDataPoints extracts DataPoints from a Prometheus query result (first series).
func extractDataPoints(result model.Value) ([]types.DataPoint, error) {
	switch v := result.(type) {
	case model.Matrix:
		if len(v) == 0 {
			return nil, nil
		}
		// Take the first series
		series := v[0]
		points := make([]types.DataPoint, len(series.Values))
		for i, sp := range series.Values {
			points[i] = types.DataPoint{
				Timestamp: sp.Timestamp.Time(),
				Value:     float64(sp.Value),
			}
		}
		return points, nil

	case model.Vector:
		if len(v) == 0 {
			return nil, nil
		}
		points := make([]types.DataPoint, len(v))
		for i, sample := range v {
			points[i] = types.DataPoint{
				Timestamp: sample.Timestamp.Time(),
				Value:     float64(sample.Value),
			}
		}
		return points, nil

	default:
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
}

// extractScalar extracts a single float64 from a Prometheus query result.
func extractScalar(result model.Value) (float64, error) {
	switch v := result.(type) {
	case model.Vector:
		if len(v) == 0 {
			return 0, fmt.Errorf("empty result")
		}
		return float64(v[0].Value), nil
	case *model.Scalar:
		return float64(v.Value), nil
	default:
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}
}

// extractMultiSeries extracts multiple series keyed by a combined label string.
func extractMultiSeries(result model.Value) (map[string][]types.DataPoint, error) {
	matrix, ok := result.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("expected matrix result, got %T", result)
	}

	seriesMap := make(map[string][]types.DataPoint)
	for _, series := range matrix {
		key := series.Metric.String()
		points := make([]types.DataPoint, len(series.Values))
		for i, sp := range series.Values {
			points[i] = types.DataPoint{
				Timestamp: sp.Timestamp.Time(),
				Value:     float64(sp.Value),
			}
		}
		seriesMap[key] = points
	}
	return seriesMap, nil
}
