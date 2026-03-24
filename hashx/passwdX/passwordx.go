package passwdX

import "golang.org/x/crypto/bcrypt"

// PasswdJiaMi cost 的取值范围是 4 到 31【默认10】，表示 2^cost 次加密循环【注意取值莫要太大影响cpu计算】。
// 原始明文密码（byte 形式）不应超过 72 字节，否则多余部分被忽略。校验时要校验这个长度
// Cost 值	迭代次数（约）	安全性	速度
// 4	16	❌ 太弱	⚡ 很快
// 10	1,024	✅ 推荐默认	🟡 适中
// 12	4,096	🔐 更安全	🔽 较慢
// 14	16,384	🔐🔐 高安全	🔻 很慢
// 31	~20亿	🤯 不现实	🐢 极慢
func PasswdJiaMi(password string, cost int) (string, error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(p), err
}

// PasswdJieMi 验证密码,srcHashedPwd为密文，dstPwd为明文【err == nil，验证完成】
func PasswdJieMi(srcHashedPwd, dstPwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(srcHashedPwd), []byte(dstPwd))
}
