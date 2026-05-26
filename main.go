package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Gagal load file .env:", err)
		return
	}

	if err := InitGlobalMinio(); err != nil {
		fmt.Println("Gagal inisialisasi MinIO:", err)
		return
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		fmt.Println("Gagal load lokasi zona waktu:", err)
		return
	}

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for {
		now := time.Now().In(loc)
		fmt.Printf("Jadwal backup: %s\n", now.Format("2006-01-02 15:04:05"))
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		if err := BackupPostgresToS3(ctx); err != nil {
			fmt.Println("Backup failed:", err)
		} else {
			fmt.Println("Backup and upload to MinIO successful.")
			
			s3KeyPrefix := strings.TrimSuffix(os.Getenv("S3_KEY_PREFIX"), "/")
			var prefix string
			if s3KeyPrefix != "" {
				prefix = s3KeyPrefix + "/"
			}
			if err := GlobalMinio.CleanOldBackups(ctx, 10, prefix); err != nil {
				fmt.Println("Failed to clean old backups:", err)
			}
		}
		cancel()
		<-ticker.C
	}
}
