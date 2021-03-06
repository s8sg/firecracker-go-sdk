package firecracker_test

import (
	"context"
	"fmt"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

func ExampleWithProcessRunner_logging() {
	const socketPath = "/tmp/firecracker.sock"
	cfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: "/path/to/kernel",
		Drives:          firecracker.NewDrivesBuilder("/path/to/rootfs").Build(),
		MachineCfg: models.MachineConfiguration{
			VcpuCount: 1,
		},
	}

	// stdout will be directed to this file
	stdoutPath := "/tmp/stdout.log"
	stdout, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to create stdout file: %v", err))
	}

	// stderr will be directed to this file
	stderrPath := "/tmp/stderr.log"
	stderr, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to create stderr file: %v", err))
	}

	ctx := context.Background()
	// build our custom command that contains our two files to
	// write to during process execution
	cmd := firecracker.VMCommandBuilder{}.
		WithBin("firecracker").
		WithSocketPath(socketPath).
		WithStdout(stdout).
		WithStderr(stderr).
		Build(ctx)

	m, err := firecracker.NewMachine(ctx, cfg, firecracker.WithProcessRunner(cmd))
	if err != nil {
		panic(fmt.Errorf("failed to create new machine: %v", err))
	}

	defer os.Remove(cfg.SocketPath)

	if err := m.Start(ctx); err != nil {
		panic(fmt.Errorf("failed to initialize machine: %v", err))
	}

	// wait for VMM to execute
	if err := m.Wait(ctx); err != nil {
		panic(err)
	}
}

func ExampleDrivesBuilder() {
	drivesParams := []struct {
		Path     string
		ReadOnly bool
	}{
		{
			Path:     "/first/path/drive.img",
			ReadOnly: true,
		},
		{
			Path:     "/second/path/drive.img",
			ReadOnly: false,
		},
	}

	// construct a new builder with the given rootfs path
	b := firecracker.NewDrivesBuilder("/path/to/rootfs")
	for _, param := range drivesParams {
		// add our additional drives
		b = b.AddDrive(param.Path, param.ReadOnly)
	}

	const socketPath = "/tmp/firecracker.sock"
	cfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: "/path/to/kernel",
		// build our drives into the machine's configuration
		Drives: b.Build(),
		MachineCfg: models.MachineConfiguration{
			VcpuCount: 1,
		},
	}

	ctx := context.Background()
	m, err := firecracker.NewMachine(ctx, cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create new machine: %v", err))
	}

	if err := m.Start(ctx); err != nil {
		panic(fmt.Errorf("failed to initialize machine: %v", err))
	}

	// wait for VMM to execute
	if err := m.Wait(ctx); err != nil {
		panic(err)
	}
}

func ExampleDrivesBuilder_DriveOpt() {
	drives := firecracker.NewDrivesBuilder("/path/to/rootfs").
		AddDrive("/path/to/drive1.img", true).
		AddDrive("/path/to/drive2.img", false, func(drive *models.Drive) {
			// set our custom bandwidth rate limiter
			drive.RateLimiter = &models.RateLimiter{
				Bandwidth: &models.TokenBucket{
					OneTimeBurst: firecracker.Int64(1024 * 1024),
					RefillTime:   firecracker.Int64(500),
					Size:         firecracker.Int64(1024 * 1024),
				},
			}
		}).
		Build()

	const socketPath = "/tmp/firecracker.sock"
	cfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: "/path/to/kernel",
		// build our drives into the machine's configuration
		Drives: drives,
		MachineCfg: models.MachineConfiguration{
			VcpuCount: 1,
		},
	}

	ctx := context.Background()
	m, err := firecracker.NewMachine(ctx, cfg)
	if err != nil {
		panic(fmt.Errorf("failed to create new machine: %v", err))
	}

	if err := m.Start(ctx); err != nil {
		panic(fmt.Errorf("failed to initialize machine: %v", err))
	}

	// wait for VMM to execute
	if err := m.Wait(ctx); err != nil {
		panic(err)
	}
}
