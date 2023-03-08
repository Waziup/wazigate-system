import React, { Component } from "react";

import {
    MDBBtn,
    MDBIcon,
    MDBCardText,
    MDBRow,
    MDBContainer,
    MDBCol
  } from "mdbreact";

class PageExport extends React.Component {
    render() {
        return (
            <>                 
            <MDBContainer>
              <MDBRow>
                <MDBCol>
                  <MDBCardText>You can export the data of all sensors and actuators to one CSV file:</MDBCardText>
                </MDBCol>
                <MDBCol>
                  <MDBBtn disabled={false} href="../../../exportall">
                        <MDBIcon
                          icon="all_match"
                          className="ml-2"
                          size="1x"
                        />{" "}
                        Export data to one CSV file
                  </MDBBtn>
                </MDBCol>
              </MDBRow>
              <MDBRow>
                <MDBCol>
                  <MDBCardText>You can export the data of all sensors and actuators to a tree of CSV files:</MDBCardText>
                </MDBCol>
                <MDBCol>
                  <MDBBtn disabled={false} href="../../../exporttree">
                        <MDBIcon
                          icon="account_tree"
                          className="ml-2"
                          size="1x"
                        />{" "}
                        Export data to tree of CSV files
                  </MDBBtn>
                </MDBCol>
              </MDBRow>
            </MDBContainer>
            </>
        );
    }   
}

export default PageExport