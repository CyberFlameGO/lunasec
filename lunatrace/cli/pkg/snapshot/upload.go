// Copyright 2022 by LunaSec (owned by Refinery Labs, Inc)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package snapshot

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"lunasec/lunatrace/pkg/graphql"
	"lunasec/lunatrace/pkg/types"
	"lunasec/lunatrace/pkg/util"
	"lunasec/lunatrace/snapshot/syftmodel"
	"net/http"
	"net/url"
)

func serializeAndCompressOutput(output syftmodel.Document) (buffer bytes.Buffer, err error) {
	serializedOutput, err := json.Marshal(output)
	if err != nil {
		log.Error().Err(err).Msg("unable to marshall dependencies output")
		return
	}

	gzwriter := gzip.NewWriter(&buffer)
	_, err = gzwriter.Write(serializedOutput)
	gzwriter.Close()

	if err != nil {
		log.Error().
			Err(err).
			Msg("Unable to compress sbom output.")
		return
	}
	return
}

func uploadCollectedSbomToUrl(
	sbomModel syftmodel.Document,
	uploadUrl string,
	uploadHeaders map[string]string,
) (err error) {
	serializedOutput, err := serializeAndCompressOutput(sbomModel)
	if err != nil {
		return
	}

	data, err := util.HttpRequest(http.MethodPut, uploadUrl, uploadHeaders, &serializedOutput)
	if err != nil {
		log.Error().
			Err(err).
			Str("response data", string(data)).
			Str("uploadUrl", uploadUrl).
			Interface("headers", uploadHeaders).
			Msg("Unable to upload SBOM data.")
		return
	}
	return
}

func getLunaTraceProjectAccessTokenHeaders(projectAccessToken string) (headers map[string]string) {
	headers = map[string]string{
		"X-LunaTrace-Access-Token": projectAccessToken,
	}
	return
}

func getOrgAndProjectFromAccessToken(
	graphqlServer types.LunaTraceGraphqlServer,
	projectAccessToken string,
) (orgId, projectId string, err error) {
	var projectInfoResponse types.GetProjectInfoResponse

	log.Debug().
		Msg("Looking up project from access token")

	headers := getLunaTraceProjectAccessTokenHeaders(projectAccessToken)

	err = graphql.PerformGraphqlRequest(
		graphqlServer,
		headers,
		graphql.NewGetProjectInfoRequest(),
		&projectInfoResponse,
	)
	if err = util.GetGraphqlError(err, projectInfoResponse.GraphqlErrors); err != nil {
		log.Error().
			Err(err).
			Msg("Unable to get project info using project secret. Make sure that your configured LUNATRACE_PROJECT_SECRET is correct.")
		return
	}

	if !projectInfoResponse.HasOnlyOneProject() {
		err = errors.New("multiple projects map to the same secret")
		log.Error().
			Err(err).
			Msg("unable to get project info, multiple projects have the same secret.")
		return
	}

	projectId = projectInfoResponse.GetProjectId()
	orgId = projectInfoResponse.GetOrganizationId()
	return
}

// todo: dry out the next 3 methods
func insertNewBuild(
	appConfig types.LunaTraceConfig,
	projectId string,
	repoMeta types.RepoMetadata,
) (agentSecret string, buildId string, err error) {
	var newBuildResponse types.NewBuildResponse

	variables := map[string]string{
		"project_id": projectId,
		"git_branch": repoMeta.BranchName,
		"git_hash":   repoMeta.CommitHash,
		"git_remote": repoMeta.RemoteUrl,
	}

	request := graphql.NewInsertNewBuildRequest(variables)

	headers := getLunaTraceProjectAccessTokenHeaders(appConfig.ProjectAccessToken)

	err = graphql.PerformGraphqlRequest(
		appConfig.GraphqlServer,
		headers,
		request,
		&newBuildResponse,
	)
	if err = util.GetGraphqlError(err, newBuildResponse.GraphqlErrors); err != nil {
		log.Error().
			Err(err).
			Interface("request", request).
			Msg("unable to create new build for project")
		return
	}
	buildId = newBuildResponse.GetBuildId()
	agentSecret = newBuildResponse.GetAgentAccessToken()
	return
}

func deleteBuild(
	appConfig types.LunaTraceConfig,
	buildId string,
) (err error) {
	var deleteBuildResponse types.DeleteBuildResponse

	request := graphql.DeleteBuildRequest(buildId)

	headers := getLunaTraceProjectAccessTokenHeaders(appConfig.ProjectAccessToken)

	err = graphql.PerformGraphqlRequest(
		appConfig.GraphqlServer,
		headers,
		request,
		&deleteBuildResponse,
	)
	if err = util.GetGraphqlError(err, deleteBuildResponse.GraphqlErrors); err != nil {
		log.Error().
			Err(err).
			Interface("request", request).
			Msg("unable to delete build")
		return
	}

	return
}

func presignSbomUpload(
	appConfig types.LunaTraceConfig,
	orgId string,
	buildId string,
) (url string, headers map[string]string, err error) {
	var presignSbomResponse types.PresignSbomResponse

	request := graphql.PresignSbomUploadRequest(orgId, buildId)

	uploadHeaders := getLunaTraceProjectAccessTokenHeaders(appConfig.ProjectAccessToken)

	err = graphql.PerformGraphqlRequest(
		appConfig.GraphqlServer,
		uploadHeaders,
		request,
		&presignSbomResponse,
	)
	if err = util.GetGraphqlError(err, presignSbomResponse.GraphqlErrors); err != nil {
		log.Error().
			Err(err).
			Interface("request", request).
			Msg("unable to fetch presigned url")
		return
	}
	url = presignSbomResponse.GetUrl()
	s3Headers := presignSbomResponse.GetHeaders()
	return url, s3Headers, err
}

func setBuildS3Url(
	appConfig types.LunaTraceConfig,
	buildId string,
	s3Url string,
) (err error) {
	var setBuildS3UrlResponse types.SetBuildS3UrlResponse

	request := graphql.UpdateBuildS3UrlRequest(buildId, s3Url)

	headers := getLunaTraceProjectAccessTokenHeaders(appConfig.ProjectAccessToken)

	err = graphql.PerformGraphqlRequest(
		appConfig.GraphqlServer,
		headers,
		request,
		&setBuildS3UrlResponse,
	)
	if err = util.GetGraphqlError(err, setBuildS3UrlResponse.GraphqlErrors); err != nil {
		log.Error().
			Err(err).
			Interface("request", request).
			Msg("unable to set s3 url for build")
		return
	}

	return
}

func uploadSbomToS3(
	appConfig types.LunaTraceConfig,
	sbomModel syftmodel.Document,
	buildId string,
	orgId string,
) (s3Url string, err error) {

	if err != nil {
		return
	}

	uploadUrl, uploadHeaders, err := presignSbomUpload(appConfig, orgId, buildId)
	if err != nil {
		return
	}

	err = uploadCollectedSbomToUrl(sbomModel, uploadUrl, uploadHeaders)
	if err != nil {
		return
	}

	s3ObjectUrl, err := url.Parse(uploadUrl)
	if err != nil {
		log.Error().
			Err(err).
			Msg("unable to parse SBOM s3 upload url")
		return
	}
	s3ObjectUrl.RawQuery = ""

	s3Url = s3ObjectUrl.String()
	return
}
