/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package hotStuff

import (
    "github.com/SealSC/SealABC/dataStructure/enum"
    "github.com/SealSC/SealABC/metadata/seal"
)

type consensusPhase struct {
    NewView     enum.Element
    Prepare     enum.Element
    PreCommit   enum.Element
    Commit      enum.Element
    Decide      enum.Element
    END         enum.Element
}

func (c consensusPhase) IntValOf(s string) (v int) {
    switch s {
    case c.NewView.String():
        return c.NewView.Int()

    case c.Prepare.String():
        return c.Prepare.Int()

    case c.PreCommit.String():
        return c.PreCommit.Int()

    case c.Commit.String():
        return c.Commit.Int()

    case c.Decide.String():
        return c.Decide.Int()

    default:
        return c.END.Int()
    }
}

var consensusPhases consensusPhase

type ConsensusPayload struct {
    Parent       []byte
    CustomerData []byte
}

type QCData struct {
    Phase      string
    ViewNumber uint64
    Payload    ConsensusPayload
}

type QC struct {
    QCData
    Votes []seal.Entity
}

type ConsensusData struct {
    ViewNumber uint64
    Phase      string
    Payload    ConsensusPayload
    Justify    QC
}

type SignedConsensusData struct {
    ConsensusData
    Seal seal.Entity
}
