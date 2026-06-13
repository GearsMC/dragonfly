package builtin

import (
	"fmt"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/i18n"
	"github.com/df-mc/dragonfly/server/performance"
	"github.com/df-mc/dragonfly/server/permission"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var registerPerformanceOnce sync.Once

// RegisterPerformance registers built-in performance inspection commands.
func RegisterPerformance() {
	registerPerformanceOnce.Do(func() {
		cmd.Register(cmd.NewWithTree("tps", i18n.D("%df.cmd.tps.description"), nil, cmd.NewCommandTree(
			cmd.Root().WithPermissions(permission.CommandTPS).Executes(tpsCommand{}),
		)))
		cmd.Register(cmd.NewWithTree("status", i18n.D("%df.cmd.status.description"), nil, cmd.NewCommandTree(
			cmd.Root().WithPermissions(permission.CommandStatus).Executes(statusCommand{}),
		)))
	})
}

type tpsCommand struct{}

func (tpsCommand) Run(src cmd.Source, o *cmd.Output, tx *world.Tx) {
	snapshot := tx.World().Metrics().Snapshot()

	o.Printm(src, "%df.performance.title")
	o.Printm(src, "%df.performance.world", snapshot.Name, snapshot.Dimension)
	o.Printm(src, "%df.performance.tps",
		healthTPS(snapshot.TPS),
		healthDuration(snapshot.Tick.Average),
		healthDuration(snapshot.Tick.Maximum),
	)
	o.Printm(src, "%df.performance.queue",
		healthQueue(src, snapshot.Queue.Current, snapshot.Queue.Peak),
		healthDuration(snapshot.Transactions.Wait.Average),
		healthDuration(snapshot.Transactions.Wait.Maximum),
	)
	o.Printm(src, "%df.performance.tx",
		healthDuration(snapshot.Transactions.Execution.Average),
		healthDuration(snapshot.Transactions.Execution.Maximum),
	)

	printOperationSummaries(src, o, snapshot.Operations)
}

type statusCommand struct{}

func (statusCommand) Run(src cmd.Source, o *cmd.Output, _ *world.Tx) {
	runtime := performance.Runtime()
	snapshots := performance.WorldSnapshots()

	o.Printm(src, "%df.status.title")
	o.Printm(src, "%df.status.memory",
		formatBytes(runtime.HeapAlloc), formatBytes(runtime.HeapInUse), runtime.HeapObjects,
	)
	o.Printm(src, "%df.status.gc",
		runtime.Goroutines, runtime.GCCycles, healthDuration(time.Duration(runtime.LastGCPause)),
	)
	o.Printm(src, "%df.status.worlds", runtime.Worlds)

	for _, snapshot := range snapshots {
		o.Printm(src, "%df.status.world.entry",
			snapshot.Name,
			snapshot.Dimension,
			healthTPS(snapshot.TPS),
			healthDuration(snapshot.Tick.Average),
			healthQueue(src, snapshot.Queue.Current, snapshot.Queue.Peak),
		)
		o.Printm(src, "%df.status.world.state",
			snapshot.State.Chunks, snapshot.State.Entities, snapshot.State.Viewers,
		)
	}
}

func printOperationSummaries(src cmd.Source, o *cmd.Output, operations map[string]performance.DurationSummary) {
	printedHeader := false
	for _, operation := range operationLabels {
		summary, ok := operations[operation.name]
		if !ok || summary.Count == 0 {
			continue
		}
		if !printedHeader {
			o.Printm(src, "%df.performance.operations")
			printedHeader = true
		}
		o.Printm(src, "%df.performance.operation",
			i18n.M(src, operation.labelKey),
			healthDuration(summary.Average),
			healthDuration(summary.Maximum),
			summary.Samples,
		)
	}
}

var operationLabels = []struct {
	name     string
	labelKey string
}{
	{name: "chunk_load", labelKey: "%df.performance.op.chunk_load"},
	{name: "chunk_generate", labelKey: "%df.performance.op.chunk_generate"},
	{name: "chunk_light_fill", labelKey: "%df.performance.op.chunk_light_fill"},
	{name: "chunk_light_spread", labelKey: "%df.performance.op.chunk_light_spread"},
	{name: "chunk_encode", labelKey: "%df.performance.op.chunk_encode"},
	{name: "subchunk_encode", labelKey: "%df.performance.op.subchunk_encode"},
	{name: "chunk_save", labelKey: "%df.performance.op.chunk_save"},
}

func healthTPS(tps float64) string {
	switch {
	case tps >= 19.5:
		return text.Green + fmt.Sprintf("%.2f", tps) + text.Reset
	case tps >= 18:
		return text.Yellow + fmt.Sprintf("%.2f", tps) + text.Reset
	default:
		return text.Red + fmt.Sprintf("%.2f", tps) + text.Reset
	}
}

func healthDuration(d time.Duration) string {
	colour := text.Green
	if d >= 50*time.Millisecond {
		colour = text.Red
	} else if d >= 25*time.Millisecond {
		colour = text.Yellow
	}
	return colour + formatDuration(d) + text.Reset
}

func healthQueue(src cmd.Source, current, peak int64) string {
	colour := text.Green
	if current >= 10 {
		colour = text.Red
	} else if current > 0 {
		colour = text.Yellow
	}
	return colour + i18n.M(src, "%df.performance.queue.status", current, peak) + text.Reset
}

func formatDuration(d time.Duration) string {
	switch {
	case d <= 0:
		return "0ms"
	case d < time.Millisecond:
		return fmt.Sprintf("%.2fus", float64(d)/float64(time.Microsecond))
	default:
		return fmt.Sprintf("%.2fms", float64(d)/float64(time.Millisecond))
	}
}

func formatBytes(bytes uint64) string {
	const mib = 1024 * 1024
	return fmt.Sprintf("%.1f MiB", float64(bytes)/mib)
}
