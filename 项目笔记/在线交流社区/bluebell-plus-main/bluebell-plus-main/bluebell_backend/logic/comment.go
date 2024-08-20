package logic

import (
	"bluebell_backend/dao/mysql"
	"bluebell_backend/models"
)

func GetCommentListByPostID(post_id string) ([]*models.CommentPlus, error) {
	// 去mysql查询数据库
	var CommentPlusList []*models.CommentPlus
	CommentList, err := mysql.GetCommentListByIDs(post_id)
	if err != nil {
		return nil, err
	}
	// 遍历返回的评论结构体列表
	for i, comment := range CommentList {
		// 查找出所有parent_id = 0的comment，表示这为一个父评论
		if comment.ParentID == 0 {
			CommentPlusList[i] = &models.CommentPlus{
				Comment: *comment,
			}
		}
	}
	// 在父节点下遍历全部节点，查找父节点的所有子节点
	for i, fcomment := range CommentPlusList {
		for _, scomment := range CommentList {
			if scomment.CommentID != 0 && fcomment.CommentID == scomment.ParentID {
				if fcomment.SubComment == nil {
					fcomment.SubComment = make([]models.Comment, 0)
				}
				fcomment.SubComment = append(fcomment.SubComment, *scomment)
			}
		}
	}
	return CommentPlusList, nil
}
