package election

import (
	"encoding/json"
	"os"
	"time"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/leaderelection"
	"k8s.io/kubernetes/pkg/client/record"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/wait"
)

const (
	startBackoff = time.Second
	maxBackoff   = time.Minute
)

type LeaderCallbacks struct {
	OnStartedLeading func(leader string)
	OnStoppedLeading func(leader string)
	OnNewLeader      func(leader string)
}

func getCurrentLeader(electionId, namespace string, c client.Interface) (string, *api.Endpoints, error) {
	endpoints, err := c.Endpoints(namespace).Get(electionId)
	if err != nil {
		return "", nil, err
	}
	val, found := endpoints.Annotations[leaderelection.LeaderElectionRecordAnnotationKey]
	if !found {
		return "", endpoints, nil
	}
	electionRecord := leaderelection.LeaderElectionRecord{}
	if err := json.Unmarshal([]byte(val), &electionRecord); err != nil {
		return "", nil, err
	}
	return electionRecord.HolderIdentity, endpoints, err
}

// NewSimpleElection creates an election, it defaults namespace to 'default' and ttl to 10s
func NewSimpleElection(electionId, id string, callback LeaderCallbacks, c client.Interface) (*leaderelection.LeaderElector, error) {
	return NewElection(electionId, id, api.NamespaceDefault, 10*time.Second, callback, c)
}

// NewElection creates an election.  'namespace'/'election' should be an existing Kubernetes Service
// 'id' is the id if this leader, should be unique.
func NewElection(electionId, id, namespace string, ttl time.Duration, callback LeaderCallbacks, c client.Interface) (*leaderelection.LeaderElector, error) {
	_, err := c.Endpoints(namespace).Get(electionId)
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.Endpoints(namespace).Create(&api.Endpoints{
				ObjectMeta: api.ObjectMeta{
					Name: electionId,
				},
			})
			if err != nil && !errors.IsConflict(err) {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	leader, endpoints, err := getCurrentLeader(electionId, namespace, c)
	if err != nil {
		return nil, err
	}
	callback.OnNewLeader(leader)

	broadcaster := record.NewBroadcaster()
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	recorder := broadcaster.NewRecorder(api.EventSource{
		Component: "leader-elector",
		Host:      hostname,
	})

	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func(stop <-chan struct{}) {
			callback.OnStartedLeading(id)
		},
		OnStoppedLeading: func() {
			leader, _, err := getCurrentLeader(electionId, namespace, c)
			if err != nil {
				glog.Errorf("failed to get leader: %v", err)
				// empty string means leader is unknown
				callback.OnStoppedLeading("")
				return
			}
			callback.OnStoppedLeading(leader)
		},
		OnNewLeader: func(identity string) {
			callback.OnNewLeader(identity)
		},
	}

	config := leaderelection.LeaderElectionConfig{
		Client:        c,
		EventRecorder: recorder,
		EndpointsMeta: endpoints.ObjectMeta,
		Identity:      id,
		LeaseDuration: ttl,
		RenewDeadline: ttl / 2,
		RetryPeriod:   ttl / 4,
		Callbacks:     callbacks,
	}

	return leaderelection.NewLeaderElector(config)
}

// RunElection runs an election given an leader elector.  Doesn't return.
func RunElection(e *leaderelection.LeaderElector) {
	wait.Forever(e.Run, 0)
}
