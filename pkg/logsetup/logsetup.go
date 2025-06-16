// Package logsetup provee una forma robusta y configurable de inicializar
// el logger global de la aplicación, con soporte para rotación de archivos y múltiples salidas.
package logsetup

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	"gopkg.in/natefinch/lumberjack.v2"
	// slogmulti "github.com/samber/slog-multi"
	// "gopkg.in/natefinch/lumberjack.v2"
)

// Config contiene todas las opciones para configurar el logger global.
type Config struct {
	Level        slog.Level // Nivel de log para la salida principal (archivo/stdout).
	ConsoleLevel slog.Level // Nivel de log específico para la salida de consola en modo "multi".
	Format       string     // Formato de salida: "text" o "json".
	Output       string     // Destino: "stdout", "stderr", "file", o "multi".
	AddSource    bool       // Incluir el archivo y línea de origen en los logs.

	// Opciones específicas para la rotación de archivos con Lumberjack.
	FileRotation struct {
		FilePath   string // Ruta al archivo de log.
		MaxSize    int    // Tamaño máximo en megabytes.
		MaxBackups int    // Número máximo de archivos de log antiguos a retener.
		MaxAge     int    // Número máximo de días a retener los logs.
		Compress   bool   // Comprimir archivos de log antiguos.
	}
}

// loggerKey es un tipo privado para usar como clave en el contexto.
type loggerKey struct{}

// Init inicializa el logger slog predeterminado para toda la aplicación.
func Init(config Config) (*slog.Logger, error) {
	var handler slog.Handler

	handlerOpts := &slog.HandlerOptions{
		AddSource: config.AddSource,
		Level:     config.Level,
	}

	if config.Output == "multi" {
		// --- Lógica para Múltiples Salidas (tu código de producción) ---
		if config.FileRotation.FilePath == "" {
			return nil, fmt.Errorf("se especificó salida 'multi' pero FileRotation.FilePath está vacío")
		}

		// 1. Configurar el logger de archivo con rotación
		logRotate := &lumberjack.Logger{
			Filename:   config.FileRotation.FilePath,
			MaxSize:    config.FileRotation.MaxSize,
			MaxBackups: config.FileRotation.MaxBackups,
			MaxAge:     config.FileRotation.MaxAge,
			Compress:   config.FileRotation.Compress,
		}
		fileHandler := slog.NewJSONHandler(logRotate, handlerOpts)

		// 2. Configurar el logger de consola
		consoleHandlerOpts := &slog.HandlerOptions{
			AddSource: config.AddSource,
			Level:     config.ConsoleLevel, // Nivel separado para la consola
		}
		consoleHandler := slog.NewTextHandler(os.Stderr, consoleHandlerOpts)

		// 3. Combinar ambos con Fanout
		handler = slogmulti.Fanout(fileHandler, consoleHandler)

	} else {
		// --- Lógica para Salida Única (nuestro código original) ---
		var writer io.Writer
		switch config.Output {
		case "stdout":
			writer = os.Stdout
		case "stderr":
			writer = os.Stderr
		case "file":
			// Nota: esto no usa rotación, para un log simple a archivo.
			// Para rotación, se debe usar el modo "multi".
			if config.FileRotation.FilePath == "" {
				return nil, fmt.Errorf("se especificó salida a archivo pero FilePath está vacío")
			}
			file, err := os.OpenFile(config.FileRotation.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				return nil, fmt.Errorf("no se pudo abrir el archivo de log %s: %w", config.FileRotation.FilePath, err)
			}
			writer = file
		default:
			return nil, fmt.Errorf("destino de output desconocido: '%s'", config.Output)
		}

		// Crear el manejador con el formato especificado
		switch config.Format {
		case "json":
			handler = slog.NewJSONHandler(writer, handlerOpts)
		case "text":
			handler = slog.NewTextHandler(writer, handlerOpts)
		default:
			return nil, fmt.Errorf("formato de log desconocido: '%s'", config.Format)
		}
	}

	// Crear el logger y establecerlo como el predeterminado global.
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Debug("Logger inicializado exitosamente", "config", fmt.Sprintf("%+v", config))
	return logger, nil
}

// SetContextWithLogger añade un logger al contexto.
func SetContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// LoggerFromContext obtiene el logger del contexto. Si no existe, devuelve el logger
// predeterminado que fue configurado por Init().
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	// Fallback seguro al logger predeterminado global.
	return slog.Default()
}
