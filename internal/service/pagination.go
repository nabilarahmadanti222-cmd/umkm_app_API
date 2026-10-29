package service

const (
    defaultLimit = 10
    maxLimit     = 100
)

// normalizePagination memastikan page minimal 1 dan limit berada pada rentang wajar.
func normalizePagination(page, limit int) (int, int) {
    if page < 1 {
        page = 1
    }
    if limit <= 0 {
        limit = defaultLimit
    }
    if limit > maxLimit {
        limit = maxLimit
    }
    return page, limit
}