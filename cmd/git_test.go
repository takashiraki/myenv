package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCloneRepository(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "test_clone")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// テスト用のリポジトリURL（小さな公開リポジトリを使用）
	testRepo := "https://github.com/octocat/Hello-World.git"
	targetPath := filepath.Join(tempDir, "test_repo")

	// CloneRepositoryをテスト
	err = CloneRepository(testRepo, targetPath)
	if err != nil {
		t.Fatalf("CloneRepository failed: %v", err)
	}

	// クローンが成功したことを確認
	if !DirIsExists(targetPath) {
		t.Errorf("Repository was not cloned to %s", targetPath)
	}

	// .gitディレクトリが存在することを確認
	gitDir := filepath.Join(targetPath, ".git")
	if !DirIsExists(gitDir) {
		t.Errorf(".git directory does not exist at %s", gitDir)
	}
}

func TestCloneRepository_InvalidRepo(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "test_clone_invalid")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 存在しないリポジトリURL
	invalidRepo := "https://github.com/nonexistent/nonexistent-repo.git"
	targetPath := filepath.Join(tempDir, "invalid_repo")

	// CloneRepositoryがエラーを返すことを確認
	err = CloneRepository(invalidRepo, targetPath)
	if err == nil {
		t.Error("Expected error for invalid repository, but got none")
	}
}

func TestCloneRepository_ExistingDirectory(t *testing.T) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "test_clone_existing")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 既存のディレクトリを作成し、ファイルを配置
	targetPath := filepath.Join(tempDir, "existing_repo")
	err = os.MkdirAll(targetPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create existing directory: %v", err)
	}
	
	// 空でないディレクトリにする
	dummyFile := filepath.Join(targetPath, "dummy.txt")
	err = os.WriteFile(dummyFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	testRepo := "https://github.com/octocat/Hello-World.git"

	// 既存のディレクトリにクローンしようとした場合のテスト
	err = CloneRepository(testRepo, targetPath)
	if err == nil {
		t.Error("Expected error for existing directory, but got none")
	}
}

// ShowLoadingIndicatorが正常に動作することを確認するテスト
func TestShowLoadingIndicatorIntegration(t *testing.T) {
	done := make(chan bool)
	
	// ゴルーチンでShowLoadingIndicatorを起動
	go ShowLoadingIndicator("テスト中", done)
	
	// 短時間待機
	time.Sleep(500 * time.Millisecond)
	
	// 停止シグナルを送信
	done <- true
	
	// テストが正常に終了すればOK（パニックしないことを確認）
}