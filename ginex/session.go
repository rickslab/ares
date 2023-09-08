package ginex

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/util"
	"github.com/sirupsen/logrus"
)

type AuthInfo struct {
	UserId int64
	Scope  string
	AppId  string
	OpenId string
}

func init() {
	gob.Register(&AuthInfo{})
}

func GetAuthInfo(c *gin.Context) *AuthInfo {
	return c.MustGet("AuthInfo").(*AuthInfo)
}

// Renewal 授权续期
func (authInfo *AuthInfo) Renewal(c *gin.Context) {
	if authInfo.Scope != "" {
		return
	}

	s := sessions.Default(c)
	s.Set("AuthInfo", authInfo)
	s.Save()
}

func (authInfo *AuthInfo) GetId() string {
	if authInfo.OpenId != "" {
		return authInfo.OpenId
	}
	return strconv.FormatInt(authInfo.UserId, 10)
}

func RedisSessionMW(redisConfKey string, sessionConfKey string) gin.HandlerFunc {
	log.Println("Use Redis-session Middleware")

	redisConf := config.YamlEnv().Sub(fmt.Sprintf("redis.%s", redisConfKey))
	sessionConf := config.YamlEnv().Sub(sessionConfKey)

	sessionKeyConf := sessionConf.GetStringSlice("session_key")
	sessionKey := make([][]byte, len(sessionKeyConf))
	for i, key := range sessionKeyConf {
		sessionKey[i] = []byte(key)
	}
	store, err := redis.NewStoreWithDB(
		10, // Pool size
		"tcp",
		redisConf.GetString("address"),
		redisConf.GetString("auth"),
		redisConf.GetString("db"),
		sessionKey..., // Codec key
	)
	util.AssertError(err)

	store.Options(sessions.Options{
		MaxAge: int(sessionConf.GetInt("session_exp") * 60 * 60), // TTL
		Path:   "/",
	})
	return sessions.Sessions("session", store)
}

func AuthorizeMW(authHandler func(c *gin.Context, accessToken string) (*AuthInfo, error)) gin.HandlerFunc {
	log.Println("Use Authorize Middleware")

	return Wrap(func(c *gin.Context) (any, error) {
		s := sessions.Default(c)
		ai := s.Get("AuthInfo")
		if ai == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return nil, nil
		}

		authInfo := ai.(*AuthInfo)
		c.Set("AuthInfo", authInfo)

		logger, ok := c.Get("Logger")
		if ok {
			c.Set("Logger", logger.(*logrus.Entry).WithField("user_id", authInfo.GetId()))
		}

		c.Next()
		return nil, nil
	})
}
