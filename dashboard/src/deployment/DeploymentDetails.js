import React, { Component } from 'react';
import { apiGet } from '../api/api.js';
import { Accordion, Header, Icon, List, Segment } from 'semantic-ui-react';
import Loading from '../util/Loading.js';
//import CommandInstruction from '../util/CommandInstruction.js';

const MemberGroupsView = ({member_groups}) => (
  <div>
    {member_groups.map((item) => <MemberListComponent 
      group={item.group}
      members={item.members}
    />)}
  </div>
);

const MemberListView = ({group, activeMemberID, onClick, members}) => (
  <Segment>
    <Header>{group}</Header>
    <List divided>
      {members.map((item) => <MemberView memberInfo={item} active={item.id === activeMemberID} onClick={onClick}/>)}
    </List>
  </Segment>
);

const MemberView = ({memberInfo, active, onClick}) => (
  <List.Item>
    <Accordion>
      <Accordion.Title active={active} onClick={() => onClick(memberInfo.id)}>
        <Icon name='dropdown' /> {memberInfo.id}
      </Accordion.Title>
      <Accordion.Content active={active}>
        <div>Pod: {memberInfo.pod_name}</div>
        <div>PVC: {memberInfo.pvc_name}</div>
        <div>PV: {memberInfo.pv_name}</div>
      </Accordion.Content>
    </Accordion>
  </List.Item>
);

class MemberListComponent extends Component {
  state = {};

  onClick = (id) => { 
    this.setState({activeMemberID:(this.state.activeMemberID === id) ? null : id}); 
  }

  render() {
    return (<MemberListView 
      group={this.props.group} 
      members={this.props.members}
      activeMemberID={this.state.activeMemberID}
      onClick={this.onClick}
    />);
  }
}

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
    const result = await apiGet(`/api/deployment/${this.props.name}`);
    this.setState({deployment:result});
  }

  render() {
    const d = this.state.deployment;
    if (!d) {
      return (<Loading/>);
    }
    return (
      <div>
        <MemberGroupsView member_groups={d.member_groups}/>
      </div>
    );
  }
}

export default DeploymentDetails;
