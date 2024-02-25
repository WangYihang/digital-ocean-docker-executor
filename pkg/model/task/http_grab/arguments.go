package http_grab

import "fmt"

type HTTPGrabArguments struct {
	InputFilePath        string
	OutputFilePath       string
	StatusUpdateFilePath string
	NumWorkers           int
	Timeout              int
	Port                 int
	Path                 string
	Host                 string
	MaxTries             int
	NumShards            int64
	Shard                int64
}

func (h *HTTPGrabArguments) String() string {
	/*
	  -i, --input=          input file path
	  -o, --output=         output file path
	  -s, --status-updates= status updates file path
	  -n, --num-workers=    number of workers (default: 32)
	  -t, --timeout=        timeout (default: 8)
	  -p, --port=           port (default: 80)
	  -P, --path=           path (default: index.html)
	  -H, --host=           host
	  -m, --max-tries=      max tries (default: 4)
	*/
	return fmt.Sprintf(
		"--input=%s --output=%s --status-updates=%s --num-workers=%d --timeout=%d --port=%d --path=%s --host=%s --max-tries=%d --num-shards=%d --shard=%d",
		h.InputFilePath,
		h.OutputFilePath,
		h.StatusUpdateFilePath,
		h.NumWorkers,
		h.Timeout,
		h.Port,
		h.Path,
		h.Host,
		h.MaxTries,
		h.NumShards,
		h.Shard,
	)
}
