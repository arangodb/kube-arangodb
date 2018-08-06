import { Loader } from 'semantic-ui-react';
import React, { Component } from 'react';
import ReactTimeout from 'react-timeout';

import { LoaderBox } from '../style/style';
import { withAuth } from '../auth/Auth.js';
import api, { isUnauthorized } from '../api/api';
import Loading from '../util/Loading';
import MemberList from './MemberList';

const MemberGroupsView = ({memberGroups, namespace}) => (
  <div>
    {memberGroups.map((item) => <MemberList 
      key={item.group}
      group={item.group}
      members={item.members}
      namespace={namespace}
    />)}
  </div>
);

class DeploymentDetails extends Component {
  state = {
    loading: true,
    error: undefined
  };

  componentDidMount() {
    this.reloadDeployment();
  }

  reloadDeployment = async() => {
    try {
      this.setState({
        loading: true
      });
      const result = await api.get(`/api/deployment/${this.props.name}`);
      this.setState({
        deployment: result,
        loading: false,
        error: undefined
      });
    } catch (e) {
      this.setState({
        loading: false,
        error: e.message
      });
      if (isUnauthorized(e)) {
        this.props.doLogout();
        return;
      }
    }
    this.props.setTimeout(this.reloadDeployment, 5000);
  }

  render() {
    const d = this.state.deployment;
    if (!d) {
      return (<Loading/>);
    }
    return (
      <div>
        <LoaderBox><Loader size="mini" active={this.state.loading} inline/></LoaderBox>
        <MemberGroupsView memberGroups={d.member_groups} namespace={d.namespace}/>
      </div>
      );
  }
}

export default ReactTimeout(withAuth(DeploymentDetails));
