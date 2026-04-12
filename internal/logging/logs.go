package logs

import (
	"log/slog"
	"os"
)

/*
exposed:
Start() - begins logging to file and handles output quality control

internals:
TODO overload funcs?
TODO formatting?
TODO Better error handling

*/

func Start() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
