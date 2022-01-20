package k8sutil

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestGetArangoDBImageIDFromPod(t *testing.T) {
	type args struct {
		pod *corev1.Pod
	}
	tests := map[string]struct {
		args    args
		want    string
		wantErr error
	}{
		"pid is nil": {
			wantErr: errors.New("failed to get container statuses from nil pod"),
		},
		"container statuses list is empty": {
			args: args{
				pod: &corev1.Pod{},
			},
			wantErr: errors.New("empty list of ContainerStatuses"),
		},
		"image ID from the only container": {
			args: args{
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						ContainerStatuses: []corev1.ContainerStatus{
							{
								ImageID: dockerPullableImageIDPrefix + "test",
							},
						},
					},
				},
			},
			want: "test",
		},
		"image ID from two containers": {
			args: args{
				pod: &corev1.Pod{
					Status: corev1.PodStatus{
						ContainerStatuses: []corev1.ContainerStatus{
							{
								ImageID: dockerPullableImageIDPrefix + "test_arango",
							},
							{
								ImageID: dockerPullableImageIDPrefix + "test1_arango",
							},
						},
					},
				},
			},
			want: "test1_arango",
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			got, err := GetArangoDBImageIDFromPod(testCase.args.pod)
			if testCase.wantErr != nil {
				require.EqualError(t, err, testCase.wantErr.Error())
				return
			}

			require.NoError(t, err)
			assert.Equalf(t, testCase.want, got, "image ID is not as expected")
		})
	}
}
