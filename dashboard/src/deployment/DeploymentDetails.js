import React, { Component } from 'react';
import api from '../api/api.js';
import Loading from '../util/Loading.js';
import MemberList from './MemberList.js';

const MemberGroupsView = ({memberGroups, namespace}) => (
  <div>
    {memberGroups.map((item) => <MemberList 
      key={`server-group-${item.group}`}
      group={item.group}
      members={item.members}
      namespace={namespace}
    />)}
  </div>
);

class DeploymentDetails extends Component {
  state = {};

  componentDidMount() {
    this.intervalId = setInterval(this.reloadDeployment, 5000);
    this.reloadDeployment();
  }

  componentWillUnmount() {
    clearInterval(this.intervalId);
  }

  reloadDeployment = async() => {
    // TODO
    const result = await api.get(`/api/deployment/${this.props.name}`);
    this.setState({deployment:result});
  }

  render() {
    const d = this.state.deployment;
    if (!d) {
      return (<Loading/>);
    }
    return (
      <div>
        <MemberGroupsView memberGroups={d.member_groups} namespace={d.namespace}/>
      </div>
    );
  }
}

export default DeploymentDetails;
