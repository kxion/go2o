/**
 * Copyright 2015 @ z3q.net.
 * name : member_summary
 * author : jarryliu
 * date : -- :
 * description :
 * history :
 */
package dto

// 会员概览信息
type MemberSummary struct {
	Id int `db:"id" auto:"yes" pk:"yes"`
	// 用户名
	Usr string `db:"usr"`
	// 昵称
	Name string `db:"name"`
	// 头像
	Avatar string `db:"avatar"`
	// 经验值
	Exp int `db:"exp"`
	// 等级
	Level int `db:"level"`
	// 等级名称
	LevelName string `db:"level_name"`
	// 等级标识
	LevelSign string `db:"program_sign"`
	// 等级是否为正式会员
	LevelOfficial int `db:"is_official"`
	// 邀请码
	InvitationCode string `db:"invitation_code"`
	// 积分
	Integral int `db:"integral"`
	// 账户余额
	Balance           float32 `db:"balance"`
	PresentBalance    float32 `db:"present_balance"`
	GrowBalance       float32 `db:"grow_balance"`
	GrowAmount        float32 `db:"grow_amount"`         // 理财总投资金额,不含收益
	GrowEarnings      float32 `db:"grow_earnings"`       // 当前收益金额
	GrowTotalEarnings float32 `db:"grow_total_earnings"` // 累积收益金额
	UpdateTime        int64   `db:"update_time"`
}
