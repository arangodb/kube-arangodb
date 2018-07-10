import React, { Component } from 'react';
import { Container, Message, Modal, Segment } from 'semantic-ui-react';

class NoOperator extends Component {
  render() {
    return (
        <Container>
          <Modal open>
            <Modal.Header>Welcome to Kube-ArangoDB</Modal.Header>
            <Modal.Content>
              <Segment basic>
                <Message color="orange">
                  There are no operators available yet.
                </Message>
              </Segment>
              {this.props.podInfoView}
              {(this.props.error) ? <Message error content={this.props.error}/> : null}
            </Modal.Content>
          </Modal>
        </Container>
    );
  }
}

export default NoOperator;
