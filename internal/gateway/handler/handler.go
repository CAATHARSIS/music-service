// internal/gateway/handler/handler.go
package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	authpb "github.com/CAATHARSIS/music-service/api/gen/auth"
	catalogpb "github.com/CAATHARSIS/music-service/api/gen/catalog"
	filepb "github.com/CAATHARSIS/music-service/api/gen/file"
	playlistpb "github.com/CAATHARSIS/music-service/api/gen/playlist"
	rulespb "github.com/CAATHARSIS/music-service/api/gen/rules"
	"github.com/CAATHARSIS/music-service/internal/gateway/config"
)

type Gateway struct {
	mux *runtime.ServeMux
	cfg *config.Config
}

func NewGateway(ctx context.Context, cfg *config.Config) (*Gateway, error) {
	mux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			return metadata.Pairs(
				"grpcgateway-http-method", req.Method,
				"grpcgateway-http-path", req.URL.Path,
			)
		}),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := catalogpb.RegisterCatalogServiceHandlerFromEndpoint(ctx, mux, cfg.CatalogServiceAddr, opts); err != nil {
		return nil, fmt.Errorf("register catalog: %w", err)
	}

	if err := authpb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, cfg.AuthServiceAddr, opts); err != nil {
		return nil, fmt.Errorf("register auth: %w", err)
	}

	if err := filepb.RegisterFileServiceHandlerFromEndpoint(ctx, mux, cfg.FileServiceAddr, opts); err != nil {
		return nil, fmt.Errorf("register file: %w", err)
	}

	if err := playlistpb.RegisterPlaylistServiceHandlerFromEndpoint(ctx, mux, cfg.PlaylistServiceAddr, opts); err != nil {
		return nil, fmt.Errorf("register playlist: %w", err)
	}

	if err := rulespb.RegisterRuleServiceHandlerFromEndpoint(ctx, mux, cfg.RulesServiceAddr, opts); err != nil {
		return nil, fmt.Errorf("register rules: %w", err)
	}

	return &Gateway{mux: mux, cfg: cfg}, nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		ctx := metadata.AppendToOutgoingContext(r.Context(), "authorization", authHeader)
		r = r.WithContext(ctx)
	}

	if r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/v1/files/upload") {
		g.handleFileUpload(w, r)
		return
	}

	g.mux.ServeHTTP(w, r)
}

func (g *Gateway) handleFileUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadedBy := ""
    if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        uploadedBy = g.extractUserID(tokenString)
    }

	conn, err := grpc.NewClient(g.cfg.FileServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := filepb.NewFileServiceClient(conn)
	stream, err := client.UploadFile(r.Context())
	if err != nil {
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	stream.Send(&filepb.UploadFileRequest{
		Data: &filepb.UploadFileRequest_Metadata{
			Metadata: &filepb.FileMetadata{
				OriginalName: header.Filename,
				MimeType:     header.Header.Get("Content-Type"),
				Bucket:       "music-files",
				UploadedBy: uploadedBy,
			},
		},
	})

	buf := make([]byte, 64<<10)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "read file failed", http.StatusInternalServerError)
			return
		}
		stream.Send(&filepb.UploadFileRequest{
			Data: &filepb.UploadFileRequest_Chunk{
				Chunk: buf[:n],
			},
		})
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		http.Error(w, fmt.Sprintf("upload failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id":"%s","original_name":"%s","size":%d}`, resp.Id, resp.OriginalName, resp.Size)
}

func (g *Gateway) extractUserID(tokenString string) string {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(g.cfg.JWTSecret), nil
	})
	if err != nil {
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	userID, _ := claims["sub"].(string)
	return userID
}
