package mysql

import (
	"context"
	v1 "github.com/marmotedu/api/apiserver/v1"
	"github.com/marmotedu/component-base/pkg/fields"
	metav1 "github.com/marmotedu/component-base/pkg/meta/v1"
	"github.com/marmotedu/errors"
	"github.com/nico612/iam-demo/internal/apiserver/store"
	"github.com/nico612/iam-demo/internal/pkg/code"
	"github.com/nico612/iam-demo/internal/pkg/util/gormutil"
	"gorm.io/gorm"
)

type users struct {
	db *gorm.DB
}

var _ store.UserStore = &users{}

func newUsers(ds *datastore) store.UserStore {
	return &users{db: ds.db}
}

// Create creates a new user account.
func (u *users) Create(ctx context.Context, user *v1.User, opts metav1.CreateOptions) error {
	return u.db.Create(&user).Error
}

// Update updates an user account information.
func (u *users) Update(ctx context.Context, user *v1.User, opts metav1.UpdateOptions) error {
	return u.db.Save(user).Error
}

// Delete deletes the user by the user identifier.
func (u *users) Delete(ctx context.Context, username string, opts metav1.DeleteOptions) error {

	// delete related policy first
	pol := newPolicies(&datastore{u.db})
	if err := pol.DeleteByUser(ctx, username, opts); err != nil {
		return err
	}

	if opts.Unscoped {
		u.db = u.db.Unscoped() // 永久删除
	}

	if err := u.db.Where("name = ?", username).Delete(&v1.User{}).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.WithCode(code.ErrDatabase, err.Error())
	}
	return nil
}

// DeleteCollection batch deletes the users.
func (u *users) DeleteCollection(ctx context.Context, usernames []string, opts metav1.DeleteOptions) error {

	// delete related policy first
	pol := newPolicies(&datastore{u.db})
	if err := pol.DeleteCollectionByUser(ctx, usernames, opts); err != nil {
		return err
	}

	if opts.Unscoped {
		u.db = u.db.Unscoped()
	}

	return u.db.Where("name in (?)").Delete(&v1.User{}).Error

}

func (u *users) Get(ctx context.Context, username string, opts metav1.GetOptions) (*v1.User, error) {

	user := &v1.User{}
	err := u.db.Where("name = ? and status = 1", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrUserNotFound, err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, err.Error())
	}

	return user, nil
}

// List return all users.
func (u *users) List(ctx context.Context, opts metav1.ListOptions) (*v1.UserList, error) {
	ret := &v1.UserList{}
	ol := gormutil.Unpointer(opts.Offset, opts.Limit)

	selector, _ := fields.ParseSelector(opts.FieldSelector)
	username, _ := selector.RequiresExactMatch("name")

	d := u.db.Where("name like ? and status = 1", "%"+username+"%").
		Offset(ol.Offset).
		Limit(ol.Limit).
		Order("id desc").
		Find(&ret.Items).
		Offset(-1).
		Limit(-1).
		Count(&ret.TotalCount)

	return ret, d.Error

}

// ListOptional show a more graceful query method.
func (u *users) ListOptional(ctx context.Context, opts metav1.ListOptions) (*v1.UserList, error) {
	ret := v1.UserList{}
	ol := gormutil.Unpointer(opts.Offset, opts.Limit)

	where := v1.User{}
	whereNot := v1.User{
		IsAdmin: 0,
	}

	selector, _ := fields.ParseSelector(opts.FieldSelector)
	username, found := selector.RequiresExactMatch("name")
	if found {
		where.Name = username
	}

	d := u.db.Where(where).Not(whereNot).Offset(ol.Offset).Limit(ol.Limit).Order("id desc").Find(&ret.Items).Count(&ret.TotalCount)
	return &ret, d.Error
}
