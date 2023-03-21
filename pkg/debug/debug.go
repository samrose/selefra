package debug

import (
	"errors"
	"github.com/natefinch/lumberjack"
	"github.com/selefra/selefra/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"
)

// SamplingService A service that samples the currently running program
type SamplingService struct {

	// The directory to which the sampled file is output
	outputDirectory string

	// Whether the system is running
	lock      sync.Mutex
	isRunning bool

	samplingInterval time.Duration

	logger *zap.Logger
}

func NewSamplingService(outputDirectory string, samplingInterval time.Duration) *SamplingService {
	_ = utils.EnsureDirectoryExists(outputDirectory)

	// init logger
	core := zapcore.NewCore(getEncoder(), getLogWriter(outputDirectory), zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller())

	return &SamplingService{
		outputDirectory:  outputDirectory,
		lock:             sync.Mutex{},
		isRunning:        false,
		samplingInterval: samplingInterval,
		logger:           logger,
	}
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(outputDirectory string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename: filepath.Join(outputDirectory, "pprof.log"),
		MaxAge:   30,
		Compress: false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func (x *SamplingService) Start() error {
	x.lock.Lock()
	defer x.lock.Unlock()

	if x.isRunning {
		return errors.New("service already running")
	}
	x.StartWorker()
	x.isRunning = true
	return nil
}

func (x *SamplingService) Stop() {
	x.lock.Lock()
	defer x.lock.Unlock()

	x.isRunning = false
}

func (x *SamplingService) IsRunning() bool {
	x.lock.Lock()
	defer x.lock.Unlock()

	return x.isRunning
}

func (x *SamplingService) StartWorker() {
	go func() {
		defer func() {
			x.logger.Debug("pprof sampling worker exit")
		}()
		for x.IsRunning() {
			x.SamplingOnce()
			time.Sleep(x.samplingInterval)
		}
	}()

}

func (x *SamplingService) SamplingOnce() {
	x.logger.Debug("begin pprof sampling...")
	begin := time.Now()
	for _, profile := range pprof.Profiles() {
		outputFilePath := filepath.Join(x.outputDirectory, profile.Name()+".pprof")
		file, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_APPEND, os.ModeAppend|os.ModePerm)
		if err != nil {
			x.logger.Error("sampling error:", zap.String("type", profile.Name()), zap.Error(err))
			continue
		}
		err = profile.WriteTo(file, 1)
		if err != nil {
			x.logger.Error("save sampling error", zap.String("type", profile.Name()), zap.Error(err))
			continue
		}
	}
	cost := time.Now().Sub(begin)
	x.logger.Debug("pprof sampling done", zap.String("cost", cost.String()))
}
