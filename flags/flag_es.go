package flags

import (
	"StarDreamerCyberNook/models"
	"StarDreamerCyberNook/service/es_service"
)

func EsIndex() {
	var tem models.ArticleModel
	es_service.Update(tem.Index(), tem.Mapping())
}
