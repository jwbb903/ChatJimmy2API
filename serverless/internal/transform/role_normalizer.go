package transform

import "github.com/taalas/chatjimmy2api/internal/types"

// normalizeRole 标准化 role 字段，确保兼容性
// 将所有已知的 role 值转换为标准格式，未知值统一转换为 assistant
func normalizeRole(role types.MessageRole) types.MessageRole {
	switch role {
	case types.RoleSystem, types.RoleUser, types.RoleAssistant, types.RoleTool:
		return role
	default:
		// 未知 role 统一转换为 assistant，避免上游服务解析失败
		return types.RoleAssistant
	}
}
