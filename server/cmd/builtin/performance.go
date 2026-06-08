package builtin

import (
	"fmt"
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/performance"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

var registerPerformanceOnce sync.Once

// RegisterPerformance registers built-in performance inspection commands.
func RegisterPerformance() {
	registerPerformanceOnce.Do(func() {
		cmd.Register(cmd.New("tps", "Displays TPS and timing metrics for the current world.", nil, tpsCommand{}))
		cmd.Register(cmd.New("status", "Displays server runtime and world performance status.", nil, statusCommand{}))
	})
}

type tpsCommand struct{}

func (tpsCommand) Run(_ cmd.Source, o *cmd.Output, tx *world.Tx) {
	snapshot := tx.World().Metrics().Snapshot()

	printColour(o, "<gold><b>=== Dunya Performansi ===</b></gold>")
	printColour(o, "<grey>Dunya:</grey> <aqua>%s</aqua> <dark-grey>(%s)</dark-grey>", snapshot.Name, snapshot.Dimension)
	printColour(o, "<grey>TPS:</grey> %s <dark-grey>|</dark-grey> <grey>Tick ortalama:</grey> %s <dark-grey>|</dark-grey> <grey>En yuksek:</grey> %s",
		healthTPS(snapshot.TPS),
		healthDuration(snapshot.Tick.Average),
		healthDuration(snapshot.Tick.Maximum),
	)
	printColour(o, "<grey>Islem kuyrugu:</grey> %s <dark-grey>|</dark-grey> <grey>Bekleme ortalama:</grey> %s <dark-grey>|</dark-grey> <grey>En yuksek:</grey> %s",
		healthQueue(snapshot.Queue.Current, snapshot.Queue.Peak),
		healthDuration(snapshot.Transactions.Wait.Average),
		healthDuration(snapshot.Transactions.Wait.Maximum),
	)
	printColour(o, "<grey>Transaction suresi:</grey> %s <dark-grey>|</dark-grey> <grey>En yuksek:</grey> %s",
		healthDuration(snapshot.Transactions.Execution.Average),
		healthDuration(snapshot.Transactions.Execution.Maximum),
	)

	printOperationSummaries(o, snapshot.Operations)
}

type statusCommand struct{}

func (statusCommand) Run(_ cmd.Source, o *cmd.Output, _ *world.Tx) {
	runtime := performance.Runtime()
	snapshots := performance.WorldSnapshots()

	printColour(o, "<gold><b>=== Sunucu Durumu ===</b></gold>")
	printColour(o, "<grey>Bellek:</grey> <aqua>%s</aqua> <dark-grey>|</dark-grey> <grey>Heap:</grey> <aqua>%s</aqua> <dark-grey>|</dark-grey> <grey>Nesne:</grey> <aqua>%d</aqua>",
		formatBytes(runtime.HeapAlloc), formatBytes(runtime.HeapInUse), runtime.HeapObjects,
	)
	printColour(o, "<grey>Gorev:</grey> <aqua>%d</aqua> <dark-grey>|</dark-grey> <grey>GC:</grey> <aqua>%d</aqua> <dark-grey>|</dark-grey> <grey>Son GC duraklamasi:</grey> %s",
		runtime.Goroutines, runtime.GCCycles, healthDuration(time.Duration(runtime.LastGCPause)),
	)
	printColour(o, "<grey>Dunyalar:</grey> <aqua>%d</aqua>", runtime.Worlds)

	for _, snapshot := range snapshots {
		printColour(o, "<yellow>%s</yellow> <dark-grey>(%s)</dark-grey> <dark-grey>-</dark-grey> TPS %s <dark-grey>|</dark-grey> Tick %s <dark-grey>|</dark-grey> Kuyruk %s",
			snapshot.Name,
			snapshot.Dimension,
			healthTPS(snapshot.TPS),
			healthDuration(snapshot.Tick.Average),
			healthQueue(snapshot.Queue.Current, snapshot.Queue.Peak),
		)
		printColour(o, "  <dark-grey>Chunk:</dark-grey> <aqua>%d</aqua> <dark-grey>| Varlik:</dark-grey> <aqua>%d</aqua> <dark-grey>| Oyuncu:</dark-grey> <aqua>%d</aqua>",
			snapshot.State.Chunks, snapshot.State.Entities, snapshot.State.Viewers,
		)
	}
}

func printOperationSummaries(o *cmd.Output, operations map[string]performance.DurationSummary) {
	printedHeader := false
	for _, operation := range operationLabels {
		summary, ok := operations[operation.name]
		if !ok || summary.Count == 0 {
			continue
		}
		if !printedHeader {
			printColour(o, "<gold><b>Olculen agir islemler</b></gold> <dark-grey>(ortalama / en yuksek)</dark-grey>")
			printedHeader = true
		}
		printColour(o, "<grey>%s:</grey> %s <dark-grey>/</dark-grey> %s <dark-grey>(%d olcum)</dark-grey>",
			operation.label,
			healthDuration(summary.Average),
			healthDuration(summary.Maximum),
			summary.Samples,
		)
	}
}

var operationLabels = []struct {
	name  string
	label string
}{
	{name: "chunk_load", label: "Chunk diskten yukleme"},
	{name: "chunk_generate", label: "Chunk uretme"},
	{name: "chunk_light_fill", label: "Ilk isiklandirma"},
	{name: "chunk_light_spread", label: "Isik yayilimi"},
	{name: "chunk_encode", label: "Chunk ag kodlama"},
	{name: "subchunk_encode", label: "Subchunk ag kodlama"},
	{name: "chunk_save", label: "Chunk kaydetme"},
}

func printColour(o *cmd.Output, format string, a ...any) {
	o.Print(text.Colourf(format, a...))
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

func healthQueue(current, peak int64) string {
	colour := text.Green
	if current >= 10 {
		colour = text.Red
	} else if current > 0 {
		colour = text.Yellow
	}
	return colour + fmt.Sprintf("simdi %d, en yuksek %d", current, peak) + text.Reset
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
