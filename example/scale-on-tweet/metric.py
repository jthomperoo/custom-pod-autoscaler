# Copyright 2019 The Custom Pod Autoscaler Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import json
import sys
import twitter

# Twitter auth environment variable keys
CONSUMER_KEY_ENV = "consumerKey"
CONSUMER_SECRET_ENV = "consumerSecret"
ACCESS_TOKEN_ENV = "accessToken"
ACCESS_TOKEN_SECRET_ENV = "accessTokenSecret"

# Hashtag environment variable key
HASHTAG_ENV = "hashtag"

def main():
    # Load twitter auth
    consumer_key = os.environ[CONSUMER_KEY_ENV]
    consumer_secret = os.environ[CONSUMER_SECRET_ENV]
    access_token = os.environ[ACCESS_TOKEN_ENV]
    access_token_secret = os.environ[ACCESS_TOKEN_SECRET_ENV]
    # Load watched hashtag
    hashtag = os.environ[HASHTAG_ENV]

    # Set up API
    api = twitter.Api(consumer_key=consumer_key,
                  consumer_secret=consumer_secret,
                  access_token_key=access_token,
                  access_token_secret=access_token_secret)

    # Get tweets with hashtag
    tweets = api.GetSearch(raw_query=f"q=%23{hashtag}&result_type=recent&count=100")

    # Count number of thumbs up and thumbs down
    num_up = 0
    num_down = 0
    for tweet in tweets:
        if "👍" in tweet.text:
            num_up += 1
            continue
        if "👎" in tweet.text:
            num_down += 1
            continue

    # Output number of thumbs up and down
    sys.stdout.write(json.dumps(
        {
            "up": num_up,
            "down": num_down
        }
    ))

if __name__ == "__main__":
    main()
