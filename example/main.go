// Copyright (c) 2019 Anbillon Team (anbillonteam@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
	_ "gitlab.com/anbillon/slago/log-to-slago"
	_ "gitlab.com/anbillon/slago/logrus-to-slago"
	"gitlab.com/anbillon/slago/slago-api"
	_ "gitlab.com/anbillon/slago/slago-zerolog"
	//_ "gitlab.com/anbillon/slago/zerolog-to-slago"
	//_ "gitlab.com/anbillon/slago/slago-logrus"
	//_ "gitlab.com/anbillon/slago/slago-zap"
	"go.uber.org/zap"
)

func main() {
	slago.Logger().AddWriter(os.Stdout)
	slago.Logger().Info().Int("int", 88).Interface("slago", "val").Msg("")
	logrus.WithField("logrus", "yes").Debug("this is from logrus")

	zap.L().Info("this is zap")

	log.Printf("this is builtin logger")
}
