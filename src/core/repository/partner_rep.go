/**
 * Copyright 2014 @ S1N1 Team.
 * name :
 * author : jarryliu
 * date : 2013-12-12 17:16
 * description :
 * history :
 */

package repository

import (
	"database/sql"
	"fmt"
	"github.com/atnet/gof/db"
	"go2o/src/core/domain/interface/member"
	"go2o/src/core/domain/interface/partner"
	"go2o/src/core/domain/interface/partner/user"
	partnerImpl "go2o/src/core/domain/partner"
	"go2o/src/core/infrastructure"
	"go2o/src/core/infrastructure/domain"
	"go2o/src/core/infrastructure/log"
	"go2o/src/core/variable"
)

var _ partner.IPartnerRep = new(partnerRep)

type partnerRep struct {
	db.Connector
	_cache     map[int]partner.IPartner
	_userRep   user.IUserRep
	_memberRep member.IMemberRep
}

func NewPartnerRep(c db.Connector, userRep user.IUserRep, memberRep member.IMemberRep) partner.IPartnerRep {
	return &partnerRep{
		Connector:  c,
		_cache:     make(map[int]partner.IPartner),
		_userRep:   userRep,
		_memberRep: memberRep,
	}
}

func (this *partnerRep) CreatePartner(v *partner.ValuePartner) (partner.IPartner, error) {
	return partnerImpl.NewPartner(v, this, this._userRep, this._memberRep)
}

func (this *partnerRep) renew(partnerId int) {
	delete(this._cache, partnerId)
}

func (this *partnerRep) GetPartner(id int) (partner.IPartner, error) {
	v, ok := this._cache[id]
	var err error
	if !ok {
		e := new(partner.ValuePartner)
		if this.Connector.GetOrm().Get(id, e) == nil {
			v, err = partnerImpl.NewPartner(e, this, this._userRep, this._memberRep)
			if v != nil {
				this._cache[id] = v
			}
		} else {
			err = partner.ErrNoSuchPartner
		}
	}
	return v, err
}

// 获取合作商主要的域名主机
func (this *partnerRep) GetPartnerMajorHost(partnerId int) string {
	//todo:
	var host string
	this.Connector.ExecScalar(`SELECT host FROM pt_siteconf WHERE partner_id=? LIMIT 0,1`,
		&host, partnerId)
	return host
}

// 保存
func (this *partnerRep) SavePartner(v *partner.ValuePartner) (int, error) {
	var err error
	if v.Id <= 0 {
		orm := this.Connector.GetOrm()
		_, _, err = orm.Save(nil, v)
		err = this.Connector.ExecScalar(`SELECT MAX(id) FROM pt_partner`, &v.Id)
		if err != nil {
			return 0, err
		}
	} else {
		_, _, err = this.Connector.GetOrm().Save(v.Id, v)
	}
	return v.Id, err
}

func (this *partnerRep) doSomething() {
	ms := []*member.ValueMember{}
	orm := this.Connector.GetOrm()
	orm.Select(&ms, "1=1")

	for _, v := range ms {
		v.Pwd = domain.Md5MemberPwd(v.Usr, "123456")
		orm.Save(v.Id, v)
	}
}

// 获取商户的编号
func (this *partnerRep) GetPartnersId() []int {

	//this.doSomething()

	dst := []int{}
	var i int

	this.Connector.Query("SELECT id FROM pt_partner", func(rows *sql.Rows) {
		for rows.Next() {
			rows.Scan(&i)
			dst = append(dst, i)
		}
		rows.Close()
	})
	return dst
}

// 获取销售配置
func (this *partnerRep) GetSaleConf(partnerId int) *partner.SaleConf {
	//10%分成
	//0.2,         #上级
	//0.1,         #上上级
	//0.8          #消费者自己
	var saleConf *partner.SaleConf = new(partner.SaleConf)
	if this.Connector.GetOrm().Get(partnerId, saleConf) == nil {
		return saleConf
	}
	return nil
}

func (this *partnerRep) SaveSaleConf(partnerId int, v *partner.SaleConf) error {
	defer this.renew(v.PartnerId)
	var err error
	if v.PartnerId > 0 {
		_, _, err = this.Connector.GetOrm().Save(v.PartnerId, v)
	} else {
		v.PartnerId = partnerId
		_, _, err = this.Connector.GetOrm().Save(nil, v)
	}
	return err
}

// 获取站点配置
func (this *partnerRep) GetSiteConf(partnerId int) *partner.SiteConf {
	var siteConf partner.SiteConf
	if err := this.Connector.GetOrm().Get(partnerId, &siteConf); err == nil {
		if len(siteConf.Host) == 0 {
			var usr string
			this.Connector.ExecScalar(
				`SELECT usr FROM pt_partner WHERE id=?`,
				&usr, partnerId)
			siteConf.Host = fmt.Sprintf("%s.%s", usr,
				infrastructure.GetApp().Config().
					GetString(variable.ServerDomain))
		}
		return &siteConf
	}
	return nil
}

func (this *partnerRep) SaveSiteConf(partnerId int, v *partner.SiteConf) error {
	defer this.renew(v.PartnerId)

	var err error
	if v.PartnerId > 0 {
		_, _, err = this.Connector.GetOrm().Save(v.PartnerId, v)
	} else {
		v.PartnerId = partnerId
		_, _, err = this.Connector.GetOrm().Save(nil, v)
	}
	return err
}

// 保存API信息
func (this *partnerRep) SaveApiInfo(partnerId int, d *partner.ApiInfo) error {
	var err error
	orm := this.GetOrm()
	if d.PartnerId == 0 { //实体未传递partnerId时新增
		d.PartnerId = partnerId
		_, _, err = orm.Save(nil, d)
	} else {
		d.PartnerId = partnerId
		_, _, err = orm.Save(partnerId, d)
	}
	return err
}

// 获取API信息
func (this *partnerRep) GetApiInfo(partnerId int) *partner.ApiInfo {
	var d *partner.ApiInfo = new(partner.ApiInfo)
	if err := this.GetOrm().Get(partnerId, d); err == nil {
		return d
	}
	return nil
}

// 根据API编号获取商户编号
func (this *partnerRep) GetPartnerIdByApiId(apiId string) int {
	var partnerId int
	this.ExecScalar("SELECT partner_id FROM pt_api WHERE api_id=?", &partnerId, apiId)
	return partnerId
}

func (this *partnerRep) SaveShop(v *partner.ValueShop) (int, error) {
	defer this.renew(v.PartnerId)
	orm := this.Connector.GetOrm()
	if v.Id > 0 {
		_, _, err := orm.Save(v.Id, v)
		return v.Id, err
	} else {
		_, _, err := orm.Save(nil, v)

		//todo: return id
		return 0, err
	}
}

func (this *partnerRep) GetValueShop(partnerId, shopId int) *partner.ValueShop {
	var v *partner.ValueShop = new(partner.ValueShop)
	err := this.Connector.GetOrm().Get(shopId, v)
	if err == nil &&
		v.PartnerId == partnerId {
		return v
	} else {
		log.PrintErr(err)
	}
	return nil
}

func (this *partnerRep) GetShopsOfPartner(partnerId int) []*partner.ValueShop {
	shops := []*partner.ValueShop{}
	err := this.Connector.GetOrm().SelectByQuery(&shops,
		"SELECT * FROM pt_shop WHERE partner_id=?", partnerId)

	if err != nil {
		log.PrintErr(err)
		return nil
	}

	return shops
}

func (this *partnerRep) DeleteShop(partnerId, shopId int) error {
	defer this.renew(partnerId)
	_, err := this.Connector.GetOrm().Delete(partner.ValueShop{},
		"partner_id=? AND id=?", partnerId, shopId)
	return err
}
