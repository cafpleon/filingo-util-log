package main

import (
	// ...
	"log"
	"log/slog"

	flng_logsetup "github.com/cafpleon/filingo-logsetup/pkg/logsetup"
)

func Log_example_main() {
	/*
		//
		// --- EJEMPLO 1: Configuración para Desarrollo ---
		// Logs de texto, en la consola, muy detallados.
		logConfigDev := flng_logsetup.Config{
			Level:     slog.LevelDebug,
			Format:    "text",
			Output:    "stderr",
			AddSource: true,
		}
	*/

	// --- EJEMPLO 2: Configuración para Producción  ---
	// Logs JSON a un archivo con rotación, y logs de nivel INFO a la consola.
	logConfigProd := flng_logsetup.Config{
		Level:        slog.LevelDebug, // Nivel para el archivo (queremos todo el detalle aquí).
		ConsoleLevel: slog.LevelInfo,  // Nivel para la consola (solo lo importante).
		Output:       "multi",         // Activa la lógica de doble salida.
		AddSource:    true,
		FileRotation: struct {
			FilePath   string
			MaxSize    int
			MaxBackups int
			MaxAge     int
			Compress   bool
		}{
			FilePath:   "log/app.log",
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		},
	}

	// Elige qué configuración usar. Podrías decidir esto basándote en una variable de entorno.
	// Por ahora, usaremos la de producción.
	if _, err := flng_logsetup.Init(logConfigProd); err != nil {
		log.Fatalf("No se pudo inicializar el logger: %v", err)
	}

	slog.Info("Aplicación iniciada con logger de producción (multi-salida)")
	slog.Debug("Este mensaje solo irá al archivo, no a la consola.")

	// ... resto de tu lógica main ...
}
