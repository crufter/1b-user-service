package users

import (
	"errors"

	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/1backend/go-utils"
	"github.com/crufter/1b-user-service/domain"
	"github.com/crufter/1b-user-service/endpoints"
	httpr "github.com/julienschmidt/httprouter"
)

type Handlers struct {
	db *gorm.DB
	ep *endpoints.Endpoints
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{
		db: db,
		ep: endpoints.NewEndpoints(db),
	}
}

func (h *Handlers) getUser(tokenId string) (*domain.User, error) {
	token, err := domain.NewAccessTokenDao(h.db).GetByToken(tokenId)
	if err != nil {
		return nil, err
	}
	u, err := domain.NewUserDao(h.db).GetById(token.UserId)
	return &u, err
}

func (h *Handlers) hasNick(tokenId, author string) error {
	token, err := domain.NewAccessTokenDao(h.db).GetByToken(tokenId)
	if err != nil {
		return err
	}
	user, err := domain.NewUserDao(h.db).GetById(token.UserId)
	if err != nil {
		return err
	}
	if user.Nick != author {
		return errors.New("No right to access")
	}
	return nil
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request, p httpr.Params) {
	eb := struct {
		Email    string
		Password string
	}{}
	if err := utils.ReadJsonBody(r, &eb); err != nil {
		utils.Write400(w, err)
		return
	}
	user, token, err := h.ep.Login(eb.Email, eb.Password)
	if err != nil {
		utils.Write400(w, err)
		return
	}
	utils.Write(w, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// We either get sent a nick, or a nick + token
// We get a token when we want to read the user by token
// We get a nick + token when either viewing our own profile or other peoples' profile
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request, p httpr.Params) {
	token := r.URL.Query().Get("token")
	nick := r.URL.Query().Get("nick")
	ownErr := h.ep.HasNick(token, nick)
	if nick == "" || ownErr == nil {
		tk, err := domain.NewAccessTokenDao(h.db).GetByToken(token)
		if err != nil {
			utils.Write400(w, err)
			return
		}
		user, err := domain.NewUserDao(h.db).GetById(tk.UserId)
		if err != nil {
			utils.Write400(w, err)
			return
		}
		utils.Write(w, user)
		return
	}
	user, err := domain.NewUserDao(h.db).GetByNick(nick)
	if err != nil {
		utils.Write400(w, err)
		return
	}
	utils.Write(w, user)
	return
}

func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request, p httpr.Params) {
	inp := struct {
		Token string
		User  domain.User
	}{}
	if err := utils.ReadJsonBody(r, &inp); err != nil {
		utils.Write400(w, err)
		return
	}
	tk, err := domain.NewAccessTokenDao(h.db).GetByToken(inp.Token)
	if err != nil {
		utils.Write400(w, err)
		return
	}
	user, err := domain.NewUserDao(h.db).GetById(tk.UserId)
	if err != nil {
		utils.Write400(w, err)
		return
	}
	user.Email = inp.User.Email
	user.AvatarLink = inp.User.AvatarLink
	user.Name = inp.User.Name
	err = domain.NewUserDao(h.db).Update(user)
	if err != nil {
		utils.Write500(w, err)
		return
	}
	utils.Write(w, map[string]string{})
}

func (h *Handlers) ChangePassword(w http.ResponseWriter, r *http.Request, p httpr.Params) {
	inp := struct {
		Token       string
		OldPassword string
		NewPassword string
	}{}
	if err := utils.ReadJsonBody(r, &inp); err != nil {
		utils.Write400(w, err)
		return
	}
	tk, err := domain.NewAccessTokenDao(h.db).GetByToken(inp.Token)
	if err != nil {
		utils.Write400(w, err)
		return
	}
	user, err := domain.NewUserDao(h.db).GetById(tk.UserId)
	if err != nil {
		utils.Write400(w, err)
		return
	}
	err = h.ep.ChangePassword(&user, inp.OldPassword, inp.NewPassword)
	if err != nil {
		utils.Write500(w, err)
		return
	}
	utils.Write(w, map[string]string{})
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request, p httpr.Params) {
	eb := struct {
		Email    string
		Password string
		Nick     string
	}{}
	if err := utils.ReadJsonBody(r, &eb); err != nil {
		utils.Write400(w, err)
		return
	}
	user, token, err := h.ep.Register(eb.Email, eb.Nick, eb.Password)
	if err != nil {
		utils.Write400(w, err)
		return
	}
	utils.Write(w, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}
