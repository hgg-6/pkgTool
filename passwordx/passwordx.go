package passwordx

import "golang.org/x/crypto/bcrypt"

// JiaMi cost 的取值范围是 4 到 31，表示 2^cost 次加密循环【注意取值莫要太大影响cpu计算】。
// 原始明文密码（byte 形式）不应超过 72 字节，否则多余部分被忽略。校验时要校验这个长度
// Cost 值	迭代次数（约）	安全性	速度
// 4	16	❌ 太弱	⚡ 很快
// 10	1,024	✅ 推荐默认	🟡 适中
// 12	4,096	🔐 更安全	🔽 较慢
// 14	16,384	🔐🔐 高安全	🔻 很慢
// 31	~20亿	🤯 不现实	🐢 极慢
func JiaMi[T string | []byte](password T, cost int) ([]byte, error) {
	var pwd []byte
	switch v := any(password).(type) {
	case string:
		pwd = []byte(v)
	case []byte:
		pwd = v
	}
	return bcrypt.GenerateFromPassword(pwd, cost)
}

// Check 验证密码,srcHashedPwd为密文，dstPwd为明文
func Check(srcHashedPwd, dstPwd string) error {
	return bcrypt.CompareHashAndPassword([]byte(srcHashedPwd), []byte(dstPwd))
}
