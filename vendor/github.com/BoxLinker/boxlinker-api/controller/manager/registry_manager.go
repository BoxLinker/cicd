package manager

import (
	"github.com/go-xorm/xorm"
	registryModels "github.com/BoxLinker/boxlinker-api/controller/models/registry"
	"github.com/BoxLinker/boxlinker-api"
)

type RegistryManager interface {
	Manager
	QueryAllACL() ([]*registryModels.ACL, error)
	SaveACL(acl *registryModels.ACL) error
	SaveImage(image *registryModels.Image) error
	ImageExistsByIdAndNamespace(id, namespace string) (bool, error)
	CountImages() (count int64, err error)
	CountImagesByNamespace(namespace string) (count int64, err error)
	QueryImages(config boxlinker.PageConfig) (images []*registryModels.Image, err error)
	QueryImagesByConditions(query interface{}, args []interface{}, cols []string, config boxlinker.PageConfig)(images []*registryModels.Image, err error)
	QueryImagesByNamespace(namespace string, config boxlinker.PageConfig) (images []*registryModels.Image, err error)
	QueryImagesByPrivilege(private bool, config boxlinker.PageConfig) (images []*registryModels.Image, err error)
	QueryImagesByNamespaceAndPrivilege(namespace string, private bool, config boxlinker.PageConfig) (images []*registryModels.Image, err error)
	FindImageById(id string) (*registryModels.Image, error)
	GetImageByIndexKey(namespace, name, tag string) (*registryModels.Image, error)

	UpdateImage(id string, image *registryModels.Image, cols ...[]string) error
	DeleteImageById(id string) error
}

type DefaultRegistryManager struct {
	DefaultManager
	engine *xorm.Engine
}

func NewRegistryManager(engine *xorm.Engine) (RegistryManager, error) {
	return &DefaultRegistryManager{
		engine: engine,
	}, nil
}

func (dm DefaultRegistryManager) ImageExistsByIdAndNamespace(id, namespace string) (bool, error) {
	var images []*registryModels.Image
	if err := dm.engine.Cols("id").Where("id = ? and namespace = ?").Find(&images); err != nil {
		return false, err
	}
	return len(images) > 0, nil
}

func (dm DefaultRegistryManager) DeleteImageById(id string) error {
	sess := dm.engine.NewSession()
	defer sess.Close()
	if _, err := dm.engine.Id(id).Delete(new(registryModels.Image)); err != nil {
		return err
	}
	return sess.Commit()
}

//
func (dm DefaultRegistryManager) UpdateImage(id string, image *registryModels.Image, cols ...[]string) error {
	sess := dm.engine.NewSession()
	defer sess.Close()
	//uImage := &registryModels.Image{}
	//uImage.Size = image.Size
	//uImage.Digest = image.Digest

	sess.ID(id)

	if len(cols) > 0{
		sess.Cols(cols[0]...)
	}

	if _, err := sess.ID(id).Update(image); err != nil {
		return err
	}
	return sess.Commit()
}

func (dm *DefaultRegistryManager) SaveImage(image *registryModels.Image) error {
	sess := dm.engine.NewSession()
	defer sess.Close()
	if _, err := sess.Insert(image); err != nil {
		return err
	}
	return sess.Commit()
}

func (dm *DefaultRegistryManager) GetImageByIndexKey(namespace, name, tag string) (*registryModels.Image, error) {
	image := &registryModels.Image{}
	has, err := dm.engine.Where("namespace = ? and name = ? and tag = ?", namespace, name, tag).Get(image)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return image, nil
}

func (dm *DefaultRegistryManager) CountImages() (count int64, err error) {
	return dm.engine.Count(new(registryModels.Image))
}
func (dm *DefaultRegistryManager) CountImagesByNamespace(namespace string) (count int64, err error) {
	return dm.engine.Where("namespace = ?", namespace).Count(new(registryModels.Image))
}

func (dm *DefaultRegistryManager) QueryImagesByConditions(query interface{}, args []interface{}, cols []string, config boxlinker.PageConfig)(images []*registryModels.Image, err error){
	sess := dm.engine.Desc("created_unix").Where(query, args...).Limit(config.Limit(), config.Offset())
	if cols != nil && len(cols) > 0 {
		sess.Cols(cols...)
	}
	err = sess.Find(&images)
	return
}

func (dm *DefaultRegistryManager) QueryImagesByNamespaceAndPrivilege(namespace string, private bool, config boxlinker.PageConfig) (images []*registryModels.Image, err error) {
	p := "0"
	if private {
		p = "1"
	}
	err = dm.engine.Desc("created_unix").Where("private = ? and namespace = ?", p, namespace).Limit(config.Limit(), config.Offset()).Find(&images)
	return
}

func (dm *DefaultRegistryManager) QueryImagesByPrivilege(private bool, config boxlinker.PageConfig) (images []*registryModels.Image, err error) {
	p := "0"
	if private {
		p = "1"
	}
	err = dm.engine.Desc("created_unix").Where("is_private = ?", p).Limit(config.Limit(), config.Offset()).Find(&images)
	return
}

func (dm *DefaultRegistryManager) QueryImagesByNamespace(namespace string, config boxlinker.PageConfig) (images []*registryModels.Image, err error) {
	err = dm.engine.Desc("created_unix").Where("namespace = ?", namespace).Limit(config.Limit(),config.Offset()).Find(&images)
	return
}

func (dm *DefaultRegistryManager) QueryImages(config boxlinker.PageConfig) (images []*registryModels.Image, err error) {
	err = dm.engine.Desc("created_unix").Limit(config.Limit(), config.Offset()).Find(&images)
	return
}

func (dm *DefaultRegistryManager) FindImageById(id string) (*registryModels.Image, error) {
	image := &registryModels.Image{}
	if _, err := dm.engine.Id(id).Get(image); err != nil {
		return nil, err
	}
	return image, nil
}



func (dm *DefaultRegistryManager) SaveACL(acl *registryModels.ACL) error {
	sess := dm.engine.NewSession()
	defer sess.Close()
	if _, err := sess.Insert(acl); err != nil {
		sess.Rollback()
		return err
	}
	return sess.Commit()
}

func (dm *DefaultRegistryManager) QueryAllACL() ([]*registryModels.ACL, error) {
	var acls []*registryModels.ACL
	if err := dm.engine.Find(&acls); err != nil {
		return nil, err
	}
	return acls, nil
}