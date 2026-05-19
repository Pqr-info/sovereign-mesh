package sovereign

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/pqr-info/sovereign-mesh/proto"
)

type ToolUseServer struct {
	proto.UnimplementedAgentToolUseServer
	browserSessions sync.Map // Map session_id to rod.Browser instance
}

func (s *ToolUseServer) ExecuteFilesystem(ctx context.Context, req *proto.FilesystemRequest) (*proto.FilesystemResponse, error) {
	switch strings.ToUpper(req.Action) {
	case "READ":
		content, err := os.ReadFile(req.Path)
		if err != nil {
			return &proto.FilesystemResponse{Success: false, Error: err.Error()}, nil
		}
		return &proto.FilesystemResponse{Content: string(content), Success: true}, nil
	case "LIST":
		files, err := os.ReadDir(req.Path)
		if err != nil {
			return &proto.FilesystemResponse{Success: false, Error: err.Error()}, nil
		}
		var fileNames []string
		for _, f := range files {
			fileNames = append(fileNames, f.Name())
		}
		return &proto.FilesystemResponse{Content: strings.Join(fileNames, "\n"), Success: true}, nil
	default:
		return &proto.FilesystemResponse{Success: false, Error: "Unsupported action"}, nil
	}
}

func (s *ToolUseServer) ExecuteWebAccess(ctx context.Context, req *proto.WebAccessRequest) (*proto.WebAccessResponse, error) {
	resp, err := http.Get(req.Url)
	if err != nil {
		return &proto.WebAccessResponse{Success: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &proto.WebAccessResponse{Success: false, Error: err.Error()}, nil
	}
	return &proto.WebAccessResponse{Html: string(body), Success: true}, nil
}

func (s *ToolUseServer) ExecuteWikipedia(ctx context.Context, req *proto.WikipediaRequest) (*proto.WikipediaResponse, error) {
	url := fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", req.Topic)
	resp, err := http.Get(url)
	if err != nil {
		return &proto.WikipediaResponse{Success: false, Error: err.Error()}, nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &proto.WikipediaResponse{Success: false, Error: err.Error()}, nil
	}
	return &proto.WikipediaResponse{Summary: string(body), Success: true}, nil
}

func (s *ToolUseServer) ExecuteBrowserAuth(ctx context.Context, req *proto.BrowserAuthRequest) (*proto.BrowserAuthResponse, error) {
	// Target the Windows Chrome profile for session persistence (WSL path)
	userDataDir := "/mnt/c/Users/theal/AppData/Local/Google/Chrome/User Data/Default"
	
	l := launcher.New().Set("user-data-dir", userDataDir).MustLaunch()
	browser := rod.New().ControlURL(l).MustConnect()
	
	page := browser.MustPage(req.TargetUrl)
	page.MustWaitLoad()
	
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	s.browserSessions.Store(sessionID, browser)
	
	return &proto.BrowserAuthResponse{Success: true, SessionToken: sessionID}, nil
}

func (s *ToolUseServer) ExecuteKeepAlive(ctx context.Context, req *proto.KeepAliveRequest) (*proto.KeepAliveResponse, error) {
	_, ok := s.browserSessions.Load(req.SessionId)
	if !ok {
		return &proto.KeepAliveResponse{Active: false, Status: "Session not found"}, nil
	}
	return &proto.KeepAliveResponse{Active: true, Status: "Session active"}, nil
}
