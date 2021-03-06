package exec

import (
	"io"
	"net/http"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubernetes/pkg/kubectl/util/term"
)

// Exec executes the command in the cnd container
func Exec(c *kubernetes.Clientset, config *rest.Config, pod *apiv1.Pod, container string, stdin io.Reader, stdout, stderr io.Writer, command []string) error {

	t := term.TTY{
		In:  stdin,
		Out: stdout,
		Raw: true,
	}

	sizeQueue := t.MonitorSize(t.GetSize())

	req := c.CoreV1().RESTClient().Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&apiv1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	fn := func() error {
		exec, err := remotecommand.NewSPDYExecutor(config, http.MethodPost, req.URL())
		if err != nil {
			return err
		}

		return exec.Stream(remotecommand.StreamOptions{
			Stdin:             stdin,
			Stdout:            stdout,
			Stderr:            stderr,
			Tty:               t.Raw,
			TerminalSizeQueue: sizeQueue,
		})
	}

	if err := t.Safe(fn); err != nil {
		return err
	}

	return nil
}
